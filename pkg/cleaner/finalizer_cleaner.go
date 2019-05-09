package cleaner

import (
	"fmt"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	sc "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog"
	"time"
)

const (
	finalizerCheckPerdiodTime = 1 * time.Second
	finalizerCheckTimeout     = 30 * time.Second
)

type FinalizerCleaner struct {
	client sc.Interface
}

// NewFinalizerCleaner returns new pointer to FinalizerCleaner
func NewFinalizerCleaner(scClient sc.Interface) *FinalizerCleaner {
	return &FinalizerCleaner{scClient}
}

// RemoveFinalizers removes specific finalizers from all ServiceCatalog CRs
func (fc *FinalizerCleaner) RemoveFinalizers() error {
	klog.V(4).Info("Removing finalizers from ClusterServiceBrokers")
	err := removeFinalizerFromClusterServiceBroker(fc.client)
	if err != nil {
		return fmt.Errorf("failed during removing ClusterServiceBroker finalizer: %s", err)
	}

	klog.V(4).Info("Removing finalizers from ServiceBrokers")
	err = removeFinalizerFromServiceBroker(fc.client)
	if err != nil {
		return fmt.Errorf("failed during removing ServiceBroker finalizer: %s", err)
	}

	klog.V(4).Info("Removing finalizers from ClusterServiceClasses")
	err = removeFinalizerFromClusterServiceClass(fc.client)
	if err != nil {
		return fmt.Errorf("failed during removing ClusterServiceClass finalizer: %s", err)
	}

	klog.V(4).Info("Removing finalizers from ServiceClasses")
	err = removeFinalizerFromServiceClass(fc.client)
	if err != nil {
		return fmt.Errorf("failed during removing ServiceClass finalizer: %s", err)
	}

	klog.V(4).Info("Removing finalizers from ClusterServicePlans")
	err = removeFinalizerFromClusterServicePlan(fc.client)
	if err != nil {
		return fmt.Errorf("failed during removing ClusterServicePlan finalizer: %s", err)
	}

	klog.V(4).Info("Removing finalizers from ServicePlans")
	err = removeFinalizerFromServicePlan(fc.client)
	if err != nil {
		return fmt.Errorf("failed during removing ServicePlan finalizer: %s", err)
	}

	klog.V(4).Info("Removing finalizers from ServiceInstances")
	err = removeFinalizerFromServiceInstance(fc.client)
	if err != nil {
		return fmt.Errorf("failed during removing ServiceInstance finalizer: %s", err)
	}

	klog.V(4).Info("Removing finalizers from ServiceBindings")
	err = removeFinalizerFromServiceBinding(fc.client)
	if err != nil {
		return fmt.Errorf("failed during removing ServiceBinding finalizer: %s", err)
	}

	return nil
}

func removeFinalizerFromClusterServiceBroker(client sc.Interface) error {
	list, err := client.ServicecatalogV1beta1().ClusterServiceBrokers().List(v1.ListOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to list ClusterServiceBrokers: %s", err)
	}

	for _, broker := range list.Items {
		finalizersList := removeServiceCatalogFinalizer(broker.Finalizers)
		toUpdate := broker.DeepCopy()
		toUpdate.Finalizers = finalizersList
		_, err := client.ServicecatalogV1beta1().ClusterServiceBrokers().Update(toUpdate)
		if err != nil {
			return fmt.Errorf("failed to update ClusterServiceBrokers %q: %s", toUpdate.Name, err)
		}
		err = wait.Poll(finalizerCheckPerdiodTime, finalizerCheckTimeout, func() (done bool, err error) {
			klog.V(4).Info("waiting for the finalizer to be removed")
			cr, err := client.ServicecatalogV1beta1().ClusterServiceBrokers().Get(toUpdate.Name, v1.GetOptions{})
			return checkFinalizerIsRemoved(cr, err)
		})
		if err != nil {
			return fmt.Errorf("failed while waiting for finalizers will be removed: %s", err)
		}
	}

	return nil
}

func removeFinalizerFromServiceBroker(client sc.Interface) error {
	list, err := client.ServicecatalogV1beta1().ServiceBrokers(v1.NamespaceAll).List(v1.ListOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to list ServiceBroker: %s", err)
	}

	for _, broker := range list.Items {
		finalizersList := removeServiceCatalogFinalizer(broker.Finalizers)
		toUpdate := broker.DeepCopy()
		toUpdate.Finalizers = finalizersList
		_, err := client.ServicecatalogV1beta1().ServiceBrokers(toUpdate.Namespace).Update(toUpdate)
		if err != nil {
			return fmt.Errorf("failed to update ServiceBroker %q: %s", toUpdate.Name, err)
		}
		err = wait.Poll(finalizerCheckPerdiodTime, finalizerCheckTimeout, func() (done bool, err error) {
			klog.V(4).Info("waiting for the finalizer to be removed")
			cr, err := client.ServicecatalogV1beta1().ServiceBrokers(toUpdate.Namespace).Get(toUpdate.Name, v1.GetOptions{})
			return checkFinalizerIsRemoved(cr, err)
		})
		if err != nil {
			return fmt.Errorf("failed while waiting for finalizers will be removed: %s", err)
		}
	}

	return nil
}

func removeFinalizerFromClusterServiceClass(client sc.Interface) error {
	list, err := client.ServicecatalogV1beta1().ClusterServiceClasses().List(v1.ListOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to list ClusterServiceClass: %s", err)
	}

	for _, class := range list.Items {
		finalizersList := removeServiceCatalogFinalizer(class.Finalizers)
		toUpdate := class.DeepCopy()
		toUpdate.Finalizers = finalizersList
		_, err := client.ServicecatalogV1beta1().ClusterServiceClasses().Update(toUpdate)
		if err != nil {
			return fmt.Errorf("failed to update ClusterServiceClass %q: %s", toUpdate.Name, err)
		}
		err = wait.Poll(finalizerCheckPerdiodTime, finalizerCheckTimeout, func() (done bool, err error) {
			klog.V(4).Info("waiting for the finalizer to be removed")
			cr, err := client.ServicecatalogV1beta1().ClusterServiceClasses().Get(toUpdate.Name, v1.GetOptions{})
			return checkFinalizerIsRemoved(cr, err)
		})
		if err != nil {
			return fmt.Errorf("failed while waiting for finalizers will be removed: %s", err)
		}
	}

	return nil
}

func removeFinalizerFromServiceClass(client sc.Interface) error {
	list, err := client.ServicecatalogV1beta1().ServiceClasses(v1.NamespaceAll).List(v1.ListOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to list ServiceClass: %s", err)
	}

	for _, class := range list.Items {
		finalizersList := removeServiceCatalogFinalizer(class.Finalizers)
		toUpdate := class.DeepCopy()
		toUpdate.Finalizers = finalizersList
		_, err := client.ServicecatalogV1beta1().ServiceClasses(toUpdate.Namespace).Update(toUpdate)
		if err != nil {
			return fmt.Errorf("failed to update ClusterServiceClass %q: %s", toUpdate.Name, err)
		}
		err = wait.Poll(finalizerCheckPerdiodTime, finalizerCheckTimeout, func() (done bool, err error) {
			klog.V(4).Info("waiting for the finalizer to be removed")
			cr, err := client.ServicecatalogV1beta1().ServiceClasses(toUpdate.Namespace).Get(toUpdate.Name, v1.GetOptions{})
			return checkFinalizerIsRemoved(cr, err)
		})
		if err != nil {
			return fmt.Errorf("failed while waiting for finalizers will be removed: %s", err)
		}
	}

	return nil
}

func removeFinalizerFromClusterServicePlan(client sc.Interface) error {
	list, err := client.ServicecatalogV1beta1().ClusterServicePlans().List(v1.ListOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to list ClusterServicePlan: %s", err)
	}

	for _, plan := range list.Items {
		finalizersList := removeServiceCatalogFinalizer(plan.Finalizers)
		toUpdate := plan.DeepCopy()
		toUpdate.Finalizers = finalizersList
		_, err := client.ServicecatalogV1beta1().ClusterServicePlans().Update(toUpdate)
		if err != nil {
			return fmt.Errorf("failed to update ClusterServicePlan %q: %s", toUpdate.Name, err)
		}
		err = wait.Poll(finalizerCheckPerdiodTime, finalizerCheckTimeout, func() (done bool, err error) {
			klog.V(4).Info("waiting for the finalizer to be removed")
			cr, err := client.ServicecatalogV1beta1().ClusterServicePlans().Get(toUpdate.Name, v1.GetOptions{})
			return checkFinalizerIsRemoved(cr, err)
		})
		if err != nil {
			return fmt.Errorf("failed while waiting for finalizers will be removed: %s", err)
		}
	}

	return nil
}

func removeFinalizerFromServicePlan(client sc.Interface) error {
	list, err := client.ServicecatalogV1beta1().ServicePlans(v1.NamespaceAll).List(v1.ListOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to list ServicePlan: %s", err)
	}

	for _, plan := range list.Items {
		finalizersList := removeServiceCatalogFinalizer(plan.Finalizers)
		toUpdate := plan.DeepCopy()
		toUpdate.Finalizers = finalizersList
		_, err := client.ServicecatalogV1beta1().ServicePlans(toUpdate.Namespace).Update(toUpdate)
		if err != nil {
			return fmt.Errorf("failed to update ServicePlan %q: %s", toUpdate.Name, err)
		}
		err = wait.Poll(finalizerCheckPerdiodTime, finalizerCheckTimeout, func() (done bool, err error) {
			klog.V(4).Info("waiting for the finalizer to be removed")
			cr, err := client.ServicecatalogV1beta1().ServicePlans(toUpdate.Namespace).Get(toUpdate.Name, v1.GetOptions{})
			return checkFinalizerIsRemoved(cr, err)
		})
		if err != nil {
			return fmt.Errorf("failed while waiting for finalizers will be removed: %s", err)
		}
	}

	return nil
}

func removeFinalizerFromServiceInstance(client sc.Interface) error {
	list, err := client.ServicecatalogV1beta1().ServiceInstances(v1.NamespaceAll).List(v1.ListOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to list ServiceInstance: %s", err)
	}

	for _, instance := range list.Items {
		finalizersList := removeServiceCatalogFinalizer(instance.Finalizers)
		toUpdate := instance.DeepCopy()
		toUpdate.Finalizers = finalizersList
		_, err := client.ServicecatalogV1beta1().ServiceInstances(toUpdate.Namespace).Update(toUpdate)
		if err != nil {
			return fmt.Errorf("failed to update ServiceInstance %q: %s", toUpdate.Name, err)
		}
		err = wait.Poll(finalizerCheckPerdiodTime, finalizerCheckTimeout, func() (done bool, err error) {
			klog.V(4).Info("waiting for the finalizer to be removed")
			cr, err := client.ServicecatalogV1beta1().ServiceInstances(toUpdate.Namespace).Get(toUpdate.Name, v1.GetOptions{})
			return checkFinalizerIsRemoved(cr, err)
		})
		if err != nil {
			return fmt.Errorf("failed while waiting for finalizers will be removed: %s", err)
		}
	}

	return nil
}

func removeFinalizerFromServiceBinding(client sc.Interface) error {
	list, err := client.ServicecatalogV1beta1().ServiceBindings(v1.NamespaceAll).List(v1.ListOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to list ServiceBinding: %s", err)
	}

	for _, binding := range list.Items {
		finalizersList := removeServiceCatalogFinalizer(binding.Finalizers)
		toUpdate := binding.DeepCopy()
		toUpdate.Finalizers = finalizersList
		_, err := client.ServicecatalogV1beta1().ServiceBindings(toUpdate.Namespace).Update(toUpdate)
		if err != nil {
			return fmt.Errorf("failed to update ServiceBinding %q: %s", toUpdate.Name, err)
		}
		err = wait.Poll(finalizerCheckPerdiodTime, finalizerCheckTimeout, func() (done bool, err error) {
			klog.V(4).Info("waiting for the finalizer to be removed")
			cr, err := client.ServicecatalogV1beta1().ServiceBindings(toUpdate.Namespace).Get(toUpdate.Name, v1.GetOptions{})
			return checkFinalizerIsRemoved(cr, err)
		})
		if err != nil {
			return fmt.Errorf("failed while waiting for finalizers will be removed: %s", err)
		}
	}

	return nil
}

type CRWithFinalizer interface {
	GetFinalizers() []string
}

func checkFinalizerIsRemoved(cr CRWithFinalizer, err error) (bool, error) {
	if errors.IsNotFound(err) {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	if len(cr.GetFinalizers()) == 0 {
		return true, nil
	}
	klog.V(4).Info("finalizers not removed, retry...")
	return false, nil
}

func removeServiceCatalogFinalizer(finalizersList []string) []string {
	finalizers := sets.NewString(finalizersList...)
	if finalizers.Has(v1beta1.FinalizerServiceCatalog) {
		finalizers.Delete(v1beta1.FinalizerServiceCatalog)
	}

	return finalizers.List()
}
