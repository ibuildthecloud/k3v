package translate

import (
	"strings"

	"github.com/rancher/wrangler/pkg/apply"
	"github.com/rancher/wrangler/pkg/kv"
	"github.com/rancher/wrangler/pkg/name"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	Generation  = "1"
	MarkerLabel = "k3v.cattle.io/managed"
)

func GetEnvVars(ip string) (result []corev1.EnvVar) {
	for _, val := range []string{
		"KUBERNETES_PORT=tcp://IP:443",
		"KUBERNETES_PORT_443_TCP=tcp://IP:443",
		"KUBERNETES_PORT_443_TCP_ADDR=IP",
		"KUBERNETES_PORT_443_TCP_PORT=443",
		"KUBERNETES_PORT_443_TCP_PROTO=tcp",
		"KUBERNETES_SERVICE_HOST=IP",
		"KUBERNETES_SERVICE_PORT=443",
		"KUBERNETES_SERVICE_PORT_HTTPS=443",
	} {
		k, v := kv.Split(val, "=")
		result = append(result, corev1.EnvVar{
			Name:  k,
			Value: strings.ReplaceAll(v, "IP", ip),
		})
	}

	return
}

func ToPName(n, ns string) string {
	return name.SafeConcatName(n, ns, "v", Generation)
}

func ObjectToPName(obj runtime.Object) string {
	meta, err := meta.Accessor(obj)
	if err != nil {
		return ""
	}
	return ToPName(meta.GetName(), meta.GetNamespace())
}

func IsManaged(obj runtime.Object) bool {
	meta, err := meta.Accessor(obj)
	if err != nil {
		return false
	}
	return meta.GetAnnotations()[MarkerLabel] == Generation
}

func GetOwner(obj runtime.Object) (namespace string, name string) {
	if !IsManaged(obj) {
		return "", ""
	}

	meta, err := meta.Accessor(obj)
	if err != nil {
		return "", ""
	}

	name = meta.GetAnnotations()[apply.LabelName]
	namespace = meta.GetAnnotations()[apply.LabelNamespace]

	return namespace, name
}

func SetupMetadata(targetNamespace string, obj runtime.Object) (runtime.Object, error) {
	target := obj.DeepCopyObject()
	if err := cpMetadata(targetNamespace, target); err != nil {
		return nil, err
	}

	return target, nil
}

func resetMostMetadata(m v1.Object) {
	// doesn't touch name, namespace, labels, annotations
	m.SetGenerateName("")
	m.SetSelfLink("")
	m.SetUID("")
	m.SetResourceVersion("")
	m.SetGeneration(0)
	m.SetCreationTimestamp(v1.Time{})
	m.SetDeletionTimestamp(nil)
	m.SetDeletionGracePeriodSeconds(nil)
	m.SetOwnerReferences(nil)
	m.SetFinalizers(nil)
	m.SetClusterName("")
	m.SetInitializers(nil)
	m.SetManagedFields(nil)
}

func cpMetadata(targetNamespace string, target runtime.Object) error {
	m, err := meta.Accessor(target)
	if err != nil {
		return err
	}

	n, ns := m.GetName(), m.GetNamespace()
	resetMostMetadata(m)

	m.SetName(ToPName(n, ns))
	m.SetNamespace(targetNamespace)

	anno := m.GetAnnotations()
	if anno == nil {
		anno = map[string]string{}
	}
	anno[MarkerLabel] = Generation
	m.SetAnnotations(anno)

	return nil
}
