package model

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/kube-openapi/pkg/util/proto"
	"k8s.io/kubectl/pkg/explain"
)

const gvkExtKey = "x-kubernetes-group-version-kind"

type Doc struct {
	schema     proto.Schema
	field      proto.Schema
	fieldsPath []string
	fieldName  string
	fieldType  string
	typeName   string
	gvk        schema.GroupVersionKind
	// schema of field ref or ref of array
	fieldRefSchema proto.Schema
}

func NewDoc(schema proto.Schema, fieldsPath []string, gvk schema.GroupVersionKind) (*Doc, error) {
	fieldName := ""
	if len(fieldsPath) != 0 {
		fieldName = fieldsPath[len(fieldsPath)-1]
	}

	field, err := explain.LookupSchemaForField(schema, fieldsPath)
	if err != nil {
		return nil, err
	}
	subSchema := findFieldSchema(field)

	typeName := explain.GetTypeName(schema)

	return &Doc{
		schema:         schema,
		field:          field,
		fieldsPath:     fieldsPath,
		fieldName:      fieldName,
		typeName:       typeName,
		gvk:            gvk,
		fieldRefSchema: subSchema,
	}, nil
}

func findFieldSchema(field proto.Schema) proto.Schema {
	if subTypeRef, ok := field.(*proto.Ref); ok {
		return subTypeRef.SubSchema()
	}
	if fieldArray, ok := field.(*proto.Array); ok {
		subType := fieldArray.SubType
		if subTypeRef, ok := subType.(*proto.Ref); ok {
			return subTypeRef.SubSchema()
		}
	}
	return nil
}

func (d *Doc) GetKind() string {
	return d.gvk.Kind
}

func (d *Doc) GetVersion() string {
	return d.gvk.Version
}

func (d *Doc) GetFieldResource() string {
	if d.fieldType == "" {
		d.fieldType = explain.GetTypeName(d.field)
	}
	if d.fieldName == "" {
		return ""
	}
	return fmt.Sprintf("%s <%s>", d.fieldName, d.fieldType)
}

func (d *Doc) GetDescriptions() []string {
	desc := []string{d.field.GetDescription()}
	if d.fieldRefSchema != nil {
		desc = append(desc, d.fieldRefSchema.GetDescription())
	}
	return desc
}

// GetDocKind returns Kind of the schema or the field ref schema
func (d *Doc) GetDocKind() *proto.Kind {
	if kind, ok := d.fieldRefSchema.(*proto.Kind); ok {
		return kind
	}
	if len(d.fieldsPath) == 0 {
		if kind, ok := d.schema.(*proto.Kind); ok {
			return kind
		}
	}
	return nil
}

func (d *Doc) FindSubDoc(fieldIdx int) *Doc {
	kind := d.GetDocKind()
	if kind == nil {
		return nil
	}

	fieldsLen := len(kind.Keys())
	if fieldIdx >= fieldsLen {
		fieldIdx = fieldsLen - 1
	}
	if fieldIdx < 0 {
		fieldIdx = 0
	}

	key := kind.Keys()[fieldIdx]

	fSchema := findFieldSchema(kind.Fields[key])
	// Field is not Ref, like string
	if fSchema == nil {
		return nil
	}

	newDoc, err := NewDoc(d.schema, append(d.fieldsPath, key), d.gvk)
	if err != nil {
		fmt.Print(err)
		return nil
	}
	return newDoc
}

func (d *Doc) FindParentDoc() *Doc {
	if len(d.fieldsPath) == 0 {
		return nil
	}
	newDoc, err := NewDoc(d.schema, d.fieldsPath[:len(d.fieldsPath)-1], d.gvk)
	if err != nil {
		fmt.Print(err)
		return d
	}
	return newDoc
}
