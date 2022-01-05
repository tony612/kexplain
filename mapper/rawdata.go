package mapper

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var singularToIndex map[string]int
var pluralToIndex map[string]int
var shortToIndex map[string]int

func init() {
	singularToIndex = make(map[string]int)
	pluralToIndex = make(map[string]int)
	shortToIndex = make(map[string]int)
	for i, item := range resources {
		gv, err := schema.ParseGroupVersion(item[2])
		if err != nil {
			continue
		}
		plural, singular := meta.UnsafeGuessKindToResource(gv.WithKind(item[3]))
		pluralToIndex[plural.Resource] = i
		singularToIndex[singular.Resource] = i
		if item[1] != "" {
			shortToIndex[item[1]] = i
		}
	}
}

// Generated by genraw
var resources = [][4]string{
	{"bindings", "", "v1", "Binding"},
	{"componentstatuses", "cs", "v1", "ComponentStatus"},
	{"configmaps", "cm", "v1", "ConfigMap"},
	{"endpoints", "ep", "v1", "Endpoints"},
	{"events", "ev", "v1", "Event"},
	{"limitranges", "limits", "v1", "LimitRange"},
	{"namespaces", "ns", "v1", "Namespace"},
	{"nodes", "no", "v1", "Node"},
	{"persistentvolumeclaims", "pvc", "v1", "PersistentVolumeClaim"},
	{"persistentvolumes", "pv", "v1", "PersistentVolume"},
	{"pods", "po", "v1", "Pod"},
	{"podtemplates", "", "v1", "PodTemplate"},
	{"replicationcontrollers", "rc", "v1", "ReplicationController"},
	{"resourcequotas", "quota", "v1", "ResourceQuota"},
	{"secrets", "", "v1", "Secret"},
	{"serviceaccounts", "sa", "v1", "ServiceAccount"},
	{"services", "svc", "v1", "Service"},
	{"mutatingwebhookconfigurations", "", "admissionregistration.k8s.io/v1", "MutatingWebhookConfiguration"},
	{"validatingwebhookconfigurations", "", "admissionregistration.k8s.io/v1", "ValidatingWebhookConfiguration"},
	{"customresourcedefinitions", "crd,crds", "apiextensions.k8s.io/v1", "CustomResourceDefinition"},
	{"apiservices", "", "apiregistration.k8s.io/v1", "APIService"},
	{"controllerrevisions", "", "apps/v1", "ControllerRevision"},
	{"daemonsets", "ds", "apps/v1", "DaemonSet"},
	{"deployments", "deploy", "apps/v1", "Deployment"},
	{"replicasets", "rs", "apps/v1", "ReplicaSet"},
	{"statefulsets", "sts", "apps/v1", "StatefulSet"},
	{"tokenreviews", "", "authentication.k8s.io/v1", "TokenReview"},
	{"localsubjectaccessreviews", "", "authorization.k8s.io/v1", "LocalSubjectAccessReview"},
	{"selfsubjectaccessreviews", "", "authorization.k8s.io/v1", "SelfSubjectAccessReview"},
	{"selfsubjectrulesreviews", "", "authorization.k8s.io/v1", "SelfSubjectRulesReview"},
	{"subjectaccessreviews", "", "authorization.k8s.io/v1", "SubjectAccessReview"},
	{"horizontalpodautoscalers", "hpa", "autoscaling/v1", "HorizontalPodAutoscaler"},
	{"cronjobs", "cj", "batch/v1beta1", "CronJob"},
	{"jobs", "", "batch/v1", "Job"},
	{"certificatesigningrequests", "csr", "certificates.k8s.io/v1", "CertificateSigningRequest"},
	{"leases", "", "coordination.k8s.io/v1", "Lease"},
	{"endpointslices", "", "discovery.k8s.io/v1beta1", "EndpointSlice"},
	{"events", "ev", "events.k8s.io/v1", "Event"},
	{"ingresses", "ing", "extensions/v1beta1", "Ingress"},
	{"ingressclasses", "", "networking.k8s.io/v1", "IngressClass"},
	{"ingresses", "ing", "networking.k8s.io/v1", "Ingress"},
	{"networkpolicies", "netpol", "networking.k8s.io/v1", "NetworkPolicy"},
	{"runtimeclasses", "", "node.k8s.io/v1beta1", "RuntimeClass"},
	{"poddisruptionbudgets", "pdb", "policy/v1beta1", "PodDisruptionBudget"},
	{"podsecuritypolicies", "psp", "policy/v1beta1", "PodSecurityPolicy"},
	{"clusterrolebindings", "", "rbac.authorization.k8s.io/v1", "ClusterRoleBinding"},
	{"clusterroles", "", "rbac.authorization.k8s.io/v1", "ClusterRole"},
	{"rolebindings", "", "rbac.authorization.k8s.io/v1", "RoleBinding"},
	{"roles", "", "rbac.authorization.k8s.io/v1", "Role"},
	{"priorityclasses", "pc", "scheduling.k8s.io/v1", "PriorityClass"},
	{"csidrivers", "", "storage.k8s.io/v1", "CSIDriver"},
	{"csinodes", "", "storage.k8s.io/v1", "CSINode"},
	{"storageclasses", "sc", "storage.k8s.io/v1", "StorageClass"},
	{"volumeattachments", "", "storage.k8s.io/v1", "VolumeAttachment"},
}
