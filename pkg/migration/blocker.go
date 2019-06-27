/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package migration

import (
	"fmt"
	"github.com/kubernetes-sigs/service-catalog/pkg/migration/blocker"
	"k8s.io/api/admissionregistration/v1beta1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/klog"
)

//DisableBlocker deletes blocking validation webhook
func (m *Service) DisableBlocker(baseName string) {
	klog.Info("Deleting deployment of WriteBlocker")

	options := metav1.DeleteOptions{}

	klog.Info("Deleting ValidatingWebhook")
	err := m.admInterface.ValidatingWebhookConfigurations().Delete(baseName, &options)
	if err != nil {
		klog.Warning(err)
	}

	klog.Info("Deleting Service")
	err = m.coreInterface.Services(m.releaseNamespace).Delete(baseName, &options)
	if err != nil {
		klog.Warning(err)
	}

	klog.Info("Deleting Pod")
	err = m.coreInterface.Pods(m.releaseNamespace).Delete(baseName, &options)
	if err != nil {
		klog.Warning(err)
	}

	klog.Info("Deleting Secret")
	err = m.coreInterface.Secrets(m.releaseNamespace).Delete(baseName+"-cert", &options)
	if err != nil {
		klog.Warning(err)
	}

	klog.Info("Deleting ServiceAccount")
	err = m.coreInterface.ServiceAccounts(m.releaseNamespace).Delete(baseName, &options)
	if err != nil {
		klog.Warning(err)

	}

	klog.Info("WriteBlocker was removed")
}

// EnableBlocker creates blocking validation webhook
func (m *Service) EnableBlocker(baseName string) error {
	klog.Info("Starting deployment of WriteBlocker")

	klog.Info("Generating Certificate")
	ca, err := blocker.GenerateCertificateAuthority(baseName+"-ca", 3650)
	if err != nil {
		return err
	}

	alternateDNS := []interface{}{
		fmt.Sprintf("%s.%s", baseName, m.releaseNamespace),
		fmt.Sprintf("%s.%s.svc", baseName, m.releaseNamespace),
	}

	cert, err := blocker.GenerateSignedCertificate(baseName, alternateDNS, 3650, ca)
	if err != nil {
		return err
	}

	klog.Info("Creating ServiceAccount")
	serviceAccount := getServiceAccountObject(baseName)
	_, err = m.coreInterface.ServiceAccounts(m.releaseNamespace).Create(serviceAccount)
	if err != nil {
		return err
	}

	klog.Info("Creating Secret")
	secret := getSecretObject(baseName, cert)
	_, err = m.coreInterface.Secrets(m.releaseNamespace).Create(secret)
	if err != nil {
		return err
	}

	klog.Info("Creating Pod")
	pod := getPodObject(baseName)
	_, err = m.coreInterface.Pods(m.releaseNamespace).Create(pod)
	if err != nil {
		return err
	}

	klog.Info("Creating Service")
	service := getServiceObject(baseName)
	_, err = m.coreInterface.Services(m.releaseNamespace).Create(service)
	if err != nil {
		return err
	}

	klog.Info("Creating ValidationWebhook")
	webhookConf := getValidationWebhookConfigurationObject(baseName, m.releaseNamespace, ca)
	_, err = m.admInterface.ValidatingWebhookConfigurations().Create(webhookConf)
	if err != nil {
		return err
	}

	klog.Info("WriteBlocker deployment finished successfully. All Service Catalog CRDs are read only")
	return nil
}

func getPodObject(name string) *core.Pod {
	return &core.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"app": name,
			},
		},
		Spec: core.PodSpec{
			Volumes: []core.Volume{
				{
					Name: name + "-cert",
					VolumeSource: core.VolumeSource{
						Secret: &core.SecretVolumeSource{
							SecretName: name + "-cert",
							Items: []core.KeyToPath{
								{
									Key:  "tls.crt",
									Path: "tls.crt",
								},
								{
									Key:  "tls.key",
									Path: "tls.key",
								},
							},
						},
					},
				},
			},
			Containers: []core.Container{
				{
					Name:            "svc",
					Image:           "awalach/service-catalog:canary",
					ImagePullPolicy: core.PullAlways,
					Args: []string{
						"migration",
						"--action",
						"start-webhook-server",
					},
					Ports: []core.ContainerPort{
						{
							ContainerPort: 8443,
							Protocol:      "TCP",
						},
					},
					VolumeMounts: []core.VolumeMount{
						{
							MountPath: "/var/run/service-catalog-blocker",
							Name:      name + "-cert",
							ReadOnly:  true,
						},
					},
				},
			},
		},
	}
}

func getServiceAccountObject(name string) *core.ServiceAccount {
	return &core.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

func getServiceObject(name string) *core.Service {
	return &core.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"app": name,
			},
		},
		Spec: core.ServiceSpec{
			Type: "NodePort",
			Selector: map[string]string{
				"app": name,
			},
			Ports: []core.ServicePort{
				{
					Name:       "secure",
					NodePort:   32443,
					Port:       443,
					Protocol:   "TCP",
					TargetPort: intstr.FromInt(8443),
				},
			},
		},
	}
}

func getSecretObject(name string, cert blocker.Certificate) *core.Secret {
	return &core.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: name + "-cert",
			Labels: map[string]string{
				"app": name,
			},
		},
		Type: "Opaque",
		Data: map[string][]byte{
			"tls.crt": []byte(cert.Cert),
			"tls.key": []byte(cert.Key),
		},
	}
}

//func getValidationWebhookConfigurationObject(name string, namespace string, cert blocker.Certificate) *v1beta1.ValidatingWebhookConfiguration {
func getValidationWebhookConfigurationObject(name string, namespace string, cert blocker.Certificate) *v1beta1.MutatingWebhookConfiguration {
	path := "/reject-changes"
	failurePolicy := v1beta1.Fail

	return &v1beta1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Webhooks: []v1beta1.Webhook{
			{
				Name:          "validating.reject-changes.servicecatalog.k8s.io",
				FailurePolicy: &failurePolicy,
				ClientConfig: v1beta1.WebhookClientConfig{
					CABundle: []byte(cert.Cert),
					Service: &v1beta1.ServiceReference{
						Name:      name,
						Namespace: namespace,
						Path:      &path,
					},
				},
				Rules: []v1beta1.RuleWithOperations{
					{
						Operations: []v1beta1.OperationType{
							v1beta1.Create,
							v1beta1.Update,
							v1beta1.Delete,
						},
						Rule: v1beta1.Rule{
							APIGroups:   []string{"servicecatalog.k8s.io"},
							APIVersions: []string{"v1beta1"},
							Resources: []string{
								"clusterservicebrokers",
								"clusterserviceclasses",
								"serviceclasses",
								"clusterserviceplans",
								"serviceplans",
								"servicebindings",
								"servicebrokers",
								"serviceinstances",
							},
						},
					},
				},
			},
		},
	}
}
