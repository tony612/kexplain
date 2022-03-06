package mapper

import "k8s.io/apimachinery/pkg/runtime/schema"

type Mapper interface {
	// Taken from k8s.io/apimachinery/pkg/api/meta.KindFor
	KindFor(resource string) (schema.GroupVersionKind, error)
}
