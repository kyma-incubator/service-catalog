/*
Copyright 2016 The Kubernetes Authors.

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

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/scheme"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

// GroupName is the group name use in this package
const GroupName = "servicecatalog.k8s.io"

// SchemeGroupVersion is group version used to register these objects
var SchemeGroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1beta1"}

// Kind takes an unqualified kind and returns a Group qualified GroupKind
func Kind(kind string) schema.GroupKind {
	return SchemeGroupVersion.WithKind(kind).GroupKind()
}

// Resource takes an unqualified resource and returns a Group qualified GroupResource
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

var (
	// SchemeBuilder needs to be exported as `SchemeBuilder` so
	// the code-generation can find it.
	SchemeBuilder      = runtime.NewSchemeBuilder(addKnownTypes, addDefaultingFuncs)
	localSchemeBuilder = &SchemeBuilder

	SchemeBuilderRuntime = &scheme.Builder{GroupVersion: SchemeGroupVersion, SchemeBuilder: SchemeBuilder}

	// AddToScheme is exposed for API installation
	AddToScheme = SchemeBuilder.AddToScheme

	//SchemeBuilderRuntime maps go types to Kubernetes GroupVersionKinds.
	SchemeBuilderRuntime = &scheme.Builder{GroupVersion: SchemeGroupVersion}
)

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&ClusterServiceBroker{},
		&ClusterServiceBrokerList{},
		&ServiceBroker{},
		&ServiceBrokerList{},
		&ClusterServiceClass{},
		&ClusterServiceClassList{},
		&ServiceClass{},
		&ServiceClassList{},
		&ClusterServicePlan{},
		&ClusterServicePlanList{},
		&ServicePlan{},
		&ServicePlanList{},
		&ServiceInstance{},
		&ServiceInstanceList{},
		&ServiceBinding{},
		&ServiceBindingList{},
	)
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	scheme.AddKnownTypes(schema.GroupVersion{Version: "v1"}, &metav1.Status{})
	scheme.AddFieldLabelConversionFunc(serviceCatalogV1Beta1GVK("ClusterServiceClass"), ClusterServiceClassFieldLabelConversionFunc)
	scheme.AddFieldLabelConversionFunc(serviceCatalogV1Beta1GVK("ServiceClass"), ServiceClassFieldLabelConversionFunc)
	scheme.AddFieldLabelConversionFunc(serviceCatalogV1Beta1GVK("ClusterServicePlan"), ClusterServicePlanFieldLabelConversionFunc)
	scheme.AddFieldLabelConversionFunc(serviceCatalogV1Beta1GVK("ServicePlan"), ServicePlanFieldLabelConversionFunc)
	scheme.AddFieldLabelConversionFunc(serviceCatalogV1Beta1GVK("ServiceInstance"), ServiceInstanceFieldLabelConversionFunc)
	scheme.AddFieldLabelConversionFunc(serviceCatalogV1Beta1GVK("ServiceBinding"), ServiceBindingFieldLabelConversionFunc)

	return nil
}

func serviceCatalogV1Beta1GVK(kind string) schema.GroupVersionKind {
	return schema.GroupVersionKind{Group: "servicecatalog.k8s.io", Version: "v1beta1", Kind: kind}
}
