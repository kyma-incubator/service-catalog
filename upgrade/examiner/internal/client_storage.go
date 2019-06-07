package internal

import (
	"fmt"
	sc "github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// ClientStorage stores all required clients in upgrade tests
type ClientStorage struct {
	client   *kubernetes.Clientset
	scClient *sc.Clientset
}

// NewClientStorage returns pointer to new ClientStorage struct
func NewClientStorage(k8sKubeconfig *rest.Config) (*ClientStorage, error) {
	clientk8s, err := kubernetes.NewForConfig(k8sKubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to get Kubernetes client: %v", err)
	}
	serviceCatalogClient, err := sc.NewForConfig(k8sKubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to get ServiceCatalog client: %v", err)
	}

	return &ClientStorage{
		client:   clientk8s,
		scClient: serviceCatalogClient,
	}, nil
}

// KubernetesClient returns kubernetes clientset
func (cs *ClientStorage) KubernetesClient() *kubernetes.Clientset {
	return cs.client
}

// ServiceCatalogClient returns ServiceCatalog clientset
func (cs *ClientStorage) ServiceCatalogClient() *sc.Clientset {
	return cs.scClient
}
