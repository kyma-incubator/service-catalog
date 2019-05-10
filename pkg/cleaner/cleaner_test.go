/*
Copyright 2018 The Kubernetes Authors.

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
	scfake "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/fake"
	"github.com/kubernetes-incubator/service-catalog/pkg/probe"
	"github.com/stretchr/testify/assert"
	"k8s.io/api/apps/v1beta1"
	extv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextfake "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	"testing"
)

const (
	cmName      = "controller-name"
	cmNamespace = "relase-namespace"
)

func TestCleaner_RemoveCRDs(t *testing.T) {
	// Given
	fakeClik8s := k8sfake.NewSimpleClientset(newTestDeployment())
	fakeCliext := apiextfake.NewSimpleClientset(newTestCRD()...)
	fakeClisc := scfake.NewSimpleClientset()

	clr := New(fakeClik8s, fakeClisc, fakeCliext)

	// When
	assert.NoError(t, clr.RemoveCRDs(cmNamespace, cmName))

	// Then
	list, err := fakeCliext.ApiextensionsV1beta1().CustomResourceDefinitions().List(v1.ListOptions{})
	assert.NoError(t, err)
	// one of the CRD is not connected with ServiceCatalog
	assert.Len(t, list.Items, 1)

	deployment, err := fakeClik8s.AppsV1beta1().Deployments(cmNamespace).Get(cmName, v1.GetOptions{})
	assert.NoError(t, err)
	assert.Equal(t, int32(0), deployment.Status.Replicas)
}

func newTestDeployment() *v1beta1.Deployment {
	var rep int32
	rep = 1
	return &v1beta1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cmName,
			Namespace: cmNamespace,
		},
		Spec: v1beta1.DeploymentSpec{
			Replicas: &rep,
		},
	}
}

func newTestCRD() []runtime.Object {
	return []runtime.Object{
		&extv1beta1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name: "NotServiceCatalogCRD",
			},
		},
		&extv1beta1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name: probe.ServiceBroker,
			},
		},
		&extv1beta1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name: probe.ClusterServiceBroker,
			},
		},
		&extv1beta1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name: probe.ServiceClass,
			},
		},
		&extv1beta1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name: probe.ClusterServiceClass,
			},
		},
		&extv1beta1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name: probe.ServicePlan,
			},
		},
		&extv1beta1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name: probe.ClusterServicePlan,
			},
		},
		&extv1beta1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name: probe.ServiceInstance,
			},
		},
		&extv1beta1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name: probe.ServiceBinding,
			},
		},
	}
}
