package server

import (
	"fmt"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
	"net/http"
)

const CRDsAmount = 8

type ReadinessCRD struct {
	client  *apiextensionsclientset.Clientset
	release string
}

// NewReadinessCRDProbe returns pointer to ReadinessCRD
func NewReadinessCRDProbe(config *rest.Config, releaseName string) (*ReadinessCRD, error) {
	apiextensionsClient, err := apiextensionsclientset.NewForConfig(config)
	if err != nil {
		return &ReadinessCRD{}, err
	}
	return &ReadinessCRD{apiextensionsClient, releaseName}, nil
}

// Name returns name of readiness probe
func (r ReadinessCRD) Name() string {
	return "ready-CRDs"
}

// Check if all CRDs with specific label are ready
func (r *ReadinessCRD) Check(_ *http.Request) error {
	list, err := r.client.ApiextensionsV1beta1().CustomResourceDefinitions().List(v1.ListOptions{
		LabelSelector: "release=" + r.release,
	})
	if err != nil {
		return fmt.Errorf("failed to list CustomResourceDefinition: %s", err)
	}
	amount := len(list.Items)
	if amount != CRDsAmount {
		return fmt.Errorf("the correct number of elements should be %d, there are %d elements", CRDsAmount, amount)
	}

	klog.V(4).Infof("Readiness probe %s checked. There are %d CRDs", r.Name(), amount)
	return nil
}
