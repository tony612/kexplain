package mapper

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type K8sMapper struct {
	mapper meta.RESTMapper
}

func NewK8sMapper(m meta.RESTMapper) *K8sMapper {
	return &K8sMapper{mapper: m}
}

func (m *K8sMapper) KindFor(resource string) (gvk schema.GroupVersionKind, err error) {
	fullySpecifiedGVR, err := m.mapper.ResourceFor(schema.GroupVersionResource{Resource: resource})
	if err != nil {
		return
	}

	gvk, _ = m.mapper.KindFor(fullySpecifiedGVR)
	if gvk.Empty() {
		gvk, err = m.mapper.KindFor(fullySpecifiedGVR.GroupResource().WithVersion(""))
		if err != nil {
			return
		}
	}
	return
}
