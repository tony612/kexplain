package mapper

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/restmapper"
)

type K8sMapper struct {
	mapper meta.RESTMapper
}

func NewK8sMapper(client discovery.DiscoveryInterface) (*K8sMapper, error) {
	apiresources, err := restmapper.GetAPIGroupResources(client)
	if err != nil {
		return nil, err
	}

	mapper := restmapper.NewDiscoveryRESTMapper(apiresources)
	mapper = restmapper.NewShortcutExpander(mapper, client)
	return &K8sMapper{mapper: mapper}, nil
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
