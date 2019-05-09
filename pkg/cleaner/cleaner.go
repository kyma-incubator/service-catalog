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
	"github.com/kubernetes-incubator/service-catalog/pkg/probe"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"time"
)

type Cleaner struct {
	client              kubernetes.Interface
	scClient            sc.Interface
	apiextensionsClient apiextensionsclientset.Interface
}

// New returns new Cleaner struct
func New(
	k8sclient kubernetes.Interface,
	serviceCatalogClient sc.Interface,
	apiExtClient apiextensionsclientset.Interface) *Cleaner {
	return &Cleaner{
		client:              k8sclient,
		scClient:            serviceCatalogClient,
		apiextensionsClient: apiExtClient,
	}
}

// RemoveCRDs takes three steps, first scale down controlle manager pod, second removes all finalizers from
// CRs and the last step removes all CRDs with specific label
func (c *Cleaner) RemoveCRDs(releaseNamespace, controllerManagerName string) error {
	err := c.scaleDownController(releaseNamespace, controllerManagerName)
	if err != nil {
		return fmt.Errorf("failed to scale down controller manager: %v", err)
	}

	err = c.removeCRDs(c.apiextensionsClient)
	if err != nil {
		return fmt.Errorf("failed to remove CustomResourceDefinitions: %v", err)
	}

	klog.V(4).Info("Removing finalizers from all ServiceCatalog custom resources")
	finalizerCleaner := NewFinalizerCleaner(c.scClient)
	err = finalizerCleaner.RemoveFinalizers()
	if err != nil {
		return fmt.Errorf("failed to remove finalizers from ServiceCatalog CRs: %s", err)
	}

	err = c.checkCRDsNotExist(c.apiextensionsClient)
	if err != nil {
		return fmt.Errorf("failed while checking CRDs not exist: %s", err)
	}

	return nil
}

func (c *Cleaner) scaleDownController(namespace, controllerName string) error {
	klog.V(4).Infof("Fetching deployment %s/%s", namespace, controllerName)
	deployment, err := c.client.
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
	_, err = c.client.AppsV1beta1().Deployments("kyma-system").Update(deploymentCopy)
	if err != nil {
		return fmt.Errorf("failed to update deployment %s/%s: %v", namespace, controllerName, err)
	}

	err = wait.Poll(3*time.Second, 120*time.Second, func() (done bool, err error) {
		klog.V(4).Info("Waiting for deployment scales down...")
		deployment, err := c.client.AppsV1().Deployments(namespace).Get(controllerName, v1.GetOptions{})
		if err != nil {
			return false, err
		}
		ready := deployment.Status.ReadyReplicas
		available := deployment.Status.AvailableReplicas
		if ready == 0 && available == 0 {
			return true, nil
		}
		klog.V(4).Info("Controller manager is not down, (ready: %d, available: %d) retry...", ready, available)
		return false, nil
	})
	if err != nil {
		return fmt.Errorf("failed during waiting for scale down controller manager: %s", err)
	}

	return nil
}

func (c *Cleaner) removeCRDs(apiextensionsClient apiextensionsclientset.Interface) error {
	klog.V(4).Info("Removing all ServiceCatalog CustomResourceDefinitions")
	list, err := apiextensionsClient.ApiextensionsV1beta1().CustomResourceDefinitions().List(v1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list CustomResourceDefinition: %s", err)
	}
	for _, crd := range list.Items {
		if !probe.IsServiceCatalogCustomResourceDefinition(crd) {
			continue
		}
		err := apiextensionsClient.ApiextensionsV1beta1().CustomResourceDefinitions().Delete(crd.Name, &v1.DeleteOptions{})
		if err != nil {
			return fmt.Errorf("failed to remove CRD %q: %s", crd.Name, err)
		}
	}

	return nil
}

func (c *Cleaner) checkCRDsNotExist(apiextensionsClient apiextensionsclientset.Interface) error {
	klog.V(4).Info("Checking all ServiceCatalog CustomResourceDefinitions are removed")
	list, err := apiextensionsClient.ApiextensionsV1beta1().CustomResourceDefinitions().List(v1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list CustomResourceDefinition: %s", err)
	}
	var amount int
	for _, crd := range list.Items {
		if probe.IsServiceCatalogCustomResourceDefinition(crd) {
			amount++
		}
	}

	if amount != 0 {
		return fmt.Errorf("CustomResourceDefinitions list is not empty. There are %s CRD(s)", amount)
	}

	return nil
}
