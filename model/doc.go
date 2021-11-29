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
	// scheuma of array items
	subSchema proto.Schema
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
	var subSchema proto.Schema = nil
	if subTypeRef, ok := field.(*proto.Ref); ok {
		subSchema = subTypeRef
	}
	if fieldArray, ok := field.(*proto.Array); ok {
		subType := fieldArray.SubType
		if subTypeRef, ok := subType.(*proto.Ref); ok {
			subSchema = subTypeRef.SubSchema()
		}
	}

	typeName := explain.GetTypeName(schema)

	return &Doc{
		schema:     schema,
		field:      field,
		fieldsPath: fieldsPath,
		fieldName:  fieldName,
		typeName:   typeName,
		gvk:        gvk,
		subSchema:  subSchema,
	}, nil
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
	if d.subSchema != nil {
		desc = append(desc, d.subSchema.GetDescription())
	}
	return desc
}

func (d *Doc) GetSubFieldKind() *proto.Kind {
	if kind, ok := d.subSchema.(*proto.Kind); ok {
		return kind
	}
	return nil
}
