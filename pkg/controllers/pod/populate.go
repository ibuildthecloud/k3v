package pod

import (
	"crypto/tls"
	"fmt"
	"io"
	ioutil2 "io/ioutil"
	"net/http"

	"github.com/rancher/k3v/pkg/translate"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	zero = int64(0)
)

func (h *handler) Remove(key string, pod *corev1.Pod) (*corev1.Pod, error) {
	return pod, h.apply.WithOwner(pod).ApplyObjects()
}

func (h *handler) Populate(key string, pod *corev1.Pod) (*corev1.Pod, error) {
	if pod == nil {
		return nil, nil
	}

	newObj, err := translate.SetupMetadata(h.targetNamespace, pod)
	if err != nil {
		return nil, err
	}

	pPod := newObj.(*corev1.Pod)
	if err := h.translatePod(pod.Namespace, pPod); err != nil {
		return pod, err
	}

	if pod.DeletionTimestamp != nil {
		h.pPods.Delete(pPod.Namespace, pPod.Name, nil)
		if pod.DeletionGracePeriodSeconds != nil && *pod.DeletionGracePeriodSeconds > 0 {
			return pod, h.vPods.Delete(pod.Namespace, pod.Name, &v1.DeleteOptions{
				GracePeriodSeconds: &zero,
			})
		}
		return pod, nil
	}

	return pod, h.apply.WithOwner(pod).ApplyObjects(newObj)
}

func (h *handler) translatePod(podNamespace string, pod *corev1.Pod) error {
	pod.Status = corev1.PodStatus{}
	pod.Spec.DeprecatedServiceAccount = ""
	pod.Spec.NodeName = ""
	pod.Spec.ServiceAccountName = ""
	pod.Spec.AutomountServiceAccountToken = &F
	pod.Spec.EnableServiceLinks = &F

	for _, v := range pod.Spec.Volumes {
		if v.ConfigMap != nil {
			v.ConfigMap.Name = translate.ToPName(v.ConfigMap.Name, podNamespace)
		}
		if v.Secret != nil {
			v.Secret.SecretName = translate.ToPName(v.Secret.SecretName, podNamespace)
		}
	}

	envVars, err := h.getEnvVars()
	if err != nil {
		return err
	}

	for i := range pod.Spec.Containers {
		for j, from := range pod.Spec.Containers[i].EnvFrom {
			if from.ConfigMapRef != nil && from.ConfigMapRef.Name != "" {
				pod.Spec.Containers[i].EnvFrom[j].ConfigMapRef.Name = translate.ToPName(pod.Spec.Containers[i].EnvFrom[j].ConfigMapRef.Name, podNamespace)
			}
			if from.SecretRef != nil && from.SecretRef.Name != "" {
				pod.Spec.Containers[i].EnvFrom[j].SecretRef.Name = translate.ToPName(pod.Spec.Containers[i].EnvFrom[j].SecretRef.Name, podNamespace)
			}
		}
		for j, env := range pod.Spec.Containers[i].Env {
			if env.ValueFrom != nil && env.ValueFrom.FieldRef != nil && env.ValueFrom.FieldRef.FieldPath == "metadata.name" {
				pod.Spec.Containers[i].Env[j].ValueFrom = nil
				pod.Spec.Containers[i].Env[j].Value = pod.Name
			}
			if env.ValueFrom != nil && env.ValueFrom.FieldRef != nil && env.ValueFrom.FieldRef.FieldPath == "metadata.namespace" {
				pod.Spec.Containers[i].Env[j].ValueFrom = nil
				pod.Spec.Containers[i].Env[j].Value = podNamespace
			}
		}
		pod.Spec.Containers[i].Env = append(pod.Spec.Containers[i].Env, envVars...)
	}

	if pod.Spec.Hostname == "" {
		if len(pod.Name) > 63 {
			pod.Spec.Hostname = pod.Name[0:63]
		} else {
			pod.Spec.Hostname = pod.Name
		}
	}

	if pod.Spec.DNSConfig == nil && (pod.Spec.DNSPolicy == corev1.DNSClusterFirst || pod.Spec.DNSPolicy == corev1.DNSClusterFirstWithHostNet) {
		nsIP, err := h.getNSIP()
		if err != nil {
			return err
		}
		five := "5"
		pod.Spec.DNSConfig = &corev1.PodDNSConfig{
			Nameservers: nsIP,
			Searches: []string{
				podNamespace + ".svc." + ClusterDomain,
				"svc." + ClusterDomain,
				ClusterDomain,
			},
			Options: []corev1.PodDNSConfigOption{
				{
					Name:  "ndots",
					Value: &five,
				},
			},
		}
		pod.Spec.DNSPolicy = corev1.DNSNone
	}

	return nil
}

func (h *handler) getEnvVars() ([]corev1.EnvVar, error) {
	k8sIP := h.getServiceIP("default", "kubernetes")
	if k8sIP == "" {
		return nil, fmt.Errorf("waiting for kubernetes service IP")
	}

	if err := h.ping(k8sIP, h.port); err != nil {
		return nil, err
	}

	return translate.GetEnvVars(k8sIP), nil
}

func (h *handler) ping(ip string, port int) error {
	h.Lock()
	defer h.Unlock()

	if h.pinged {
		return nil
	}

	c := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	defer c.CloseIdleConnections()

	req, err := http.NewRequest("GET", fmt.Sprintf("https://localhost:%d", port), nil)
	if err != nil {
		return err
	}
	req.Host = ip
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	io.Copy(ioutil2.Discard, resp.Body)
	resp.Body.Close()

	h.pinged = true
	return nil
}

func (h *handler) getNSIP() ([]string, error) {
	ip := h.getServiceIP("kube-system", "kube-dns")
	if ip == "" {
		return nil, fmt.Errorf("waiting for DNS service IP")
	}
	return []string{ip}, nil
}

func (h *handler) getServiceIP(namespace, name string) string {
	pName := translate.ToPName(name, namespace)
	svc, err := h.pServiceCache.Get(h.targetNamespace, pName)
	if err != nil {
		return ""
	}

	return svc.Spec.ClusterIP
}
