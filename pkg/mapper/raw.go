package mapper

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

type RawMapper struct{}

func NewRawMapper() *RawMapper {
	return &RawMapper{}
}

func (r *RawMapper) KindFor(resource string) (schema.GroupVersionKind, error) {
	idx := -1
	if i, ok := singularToIndex[resource]; ok {
		idx = i
	}
	if i, ok := pluralToIndex[resource]; ok {
		idx = i
	}
	if i, ok := shortToIndex[resource]; ok {
		idx = i
	}
	if idx < 0 {
		return schema.GroupVersionKind{}, fmt.Errorf("not found kind for %s using raw mapper", resource)
	}

	item := resources[idx]
	gv, err := schema.ParseGroupVersion(item[2])
	if err != nil {
		return schema.GroupVersionKind{}, err
	}
	return gv.WithKind(item[3]), nil
}
