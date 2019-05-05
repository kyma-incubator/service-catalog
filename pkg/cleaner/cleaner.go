/*
Copyright 2017 The Kubernetes Authors.

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

package cleaner

import (
	"fmt"
	sc "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
	"time"
)

type Cleaner struct {
	config *rest.Config
}

// NewCleaner returns new Cleaner struct
func NewCleaner() (*Cleaner, error) {
	k8sKubeconfig, err := rest.InClusterConfig()
	if err != nil {
		return &Cleaner{}, fmt.Errorf("failed to get Kubernetes client config: %v", err)
	}

	return &Cleaner{k8sKubeconfig}, nil
}

// RemoveCRDs takes three steps, first scale down controlle manager pod, second removes all finalizers from
// CRs and the last step removes all CRDs with specific label
func (c *Cleaner) RemoveCRDs(releaseName, releaseNamespace, controllerManagerName string) error {
	err := c.scaleDownController(releaseNamespace, controllerManagerName)
	if err != nil {
		return fmt.Errorf("failed to scale down controller manager: %v", err)
	}

	scClient, err := sc.NewForConfig(c.config)
	if err != nil {
		return fmt.Errorf("failed to get ServiceCatalog client: %v", err)
	}
	apiextensionsClient, err := apiextensionsclientset.NewForConfig(c.config)
	if err != nil {
		return fmt.Errorf("failed to get Apiextensions client: %v", err)
	}

	klog.V(4).Info("Removing finalizers from all ServiceCatalog custom resources")
	finalizerCleaner := NewFinalizerCleaner(scClient)
	err = finalizerCleaner.RemoveFinalizers()
	if err != nil {
		return fmt.Errorf("failed to remove finalizers from ServiceCatalog CRs: %s", err)
	}

	klog.V(4).Info("Removing all ServiceCatalog CustomResourceDefinitions")
	list, err := apiextensionsClient.ApiextensionsV1beta1().CustomResourceDefinitions().List(v1.ListOptions{
		LabelSelector: "release=" + releaseName,
	})
	if err != nil {
		return fmt.Errorf("failed to list CustomResourceDefinition: %s", err)
	}
	for _, crd := range list.Items {
		err := apiextensionsClient.ApiextensionsV1beta1().CustomResourceDefinitions().Delete(crd.Name, &v1.DeleteOptions{})
		if err != nil {
			return fmt.Errorf("failed to remove CRD %q: %s", crd.Name, err)
		}
	}

	return nil
}

func (c *Cleaner) scaleDownController(namespace, controllerName string) error {
	client, err := kubernetes.NewForConfig(c.config)
	if err != nil {
		return fmt.Errorf("failed to get kubernetes client: %s", err)
	}

	klog.V(4).Infof("Fetching deployment %s/%s", namespace, controllerName)
	deployment, err := client.
		AppsV1beta1().
		Deployments(namespace).
		Get(controllerName, v1.GetOptions{})
	if err != nil {
		return fmt.Errorf("cannot get deployment %s/%s: %s", namespace, controllerName, err)
	}

	klog.V(4).Info("Scaling down deployment to zero")
	replicas := int32(0)
	deploymentCopy := deployment.DeepCopy()
	deploymentCopy.Spec.Replicas = &replicas
	_, err = client.AppsV1beta1().Deployments("kyma-system").Update(deploymentCopy)
	if err != nil {
		return fmt.Errorf("failed to update deployment $s/%s: %v", namespace, controllerName, err)
	}

	err = wait.Poll(3*time.Second, 120*time.Second, func() (done bool, err error) {
		klog.V(4).Info("Waiting for pods to be removed...")
		podList, err := client.CoreV1().Pods(namespace).List(v1.ListOptions{LabelSelector: "app=" + controllerName})
		if err != nil {
			return false, err
		}
		if len(podList.Items) == 0 {
			return true, nil
		}
		klog.V(4).Info("Controller manager pods are not down. There are %d pods, retry...", len(podList.Items))
		return false, nil
	})
	if err != nil {
		return fmt.Errorf("failed during waiting for scale down controller manager pods: %s", err)
	}

	return nil
}
