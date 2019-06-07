package cluster_service_broker

import (
	scClientset "github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	"github.com/pkg/errors"
	apiErr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog"
)

type cleaner struct {
	sc        scClientset.ServicecatalogV1beta1Interface
	namespace string
}

func newCleaner(cli ClientGetter, ns string) *cleaner {
	return &cleaner{
		sc:        cli.ServiceCatalogClient().ServicecatalogV1beta1(),
		namespace: ns,
	}
}

func (c *cleaner) clean() error {
	klog.Info("Start cleaning resources for ClusterServiceBroker test")
	for _, fn := range []func() error{
		c.removeServiceBinding,
		c.removeServiceInstance,
		c.unregisterClusterServiceBroker,
	} {
		err := fn()
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *cleaner) removeServiceBinding() error {
	exist, err := c.serviceBindingExist()
	if err != nil {
		return errors.Wrap(err, "failed during fetching ServiceBinding")
	}
	if !exist {
		return nil
	}
	if err := c.deleteServiceBinding(); err != nil {
		return errors.Wrap(err, "failed during removing ServiceBinding")
	}
	if err := c.assertServiceBindingIsRemoved(); err != nil {
		return errors.Wrap(err, "failed during asserting ServiceBinding is removed")
	}
	return nil
}

func (c *cleaner) removeServiceInstance() error {
	exist, err := c.serviceInstanceExist()
	if err != nil {
		return errors.Wrap(err, "failed during fetching ServiceInstance")
	}
	if !exist {
		return nil
	}
	// remove `removeServiceInstanceFinalizer` method if BrokerTest will be fixed and
	// will handle ServiceInstance delete operation
	// for now BrokerTest failed and ServiceInstance has deprovisioning false status
	if err := c.removeServiceInstanceFinalizer(); err != nil {
		return errors.Wrap(err, "failed during removing ServiceInstance finalizers")
	}
	if err := c.deleteServiceInstance(); err != nil {
		return errors.Wrap(err, "failed during removing ServiceInstance")
	}
	if err := c.assertServiceInstanceIsRemoved(); err != nil {
		return errors.Wrap(err, "failed during asserting ServiceInstance is removed")
	}
	return nil
}

func (c *cleaner) unregisterClusterServiceBroker() error {
	if err := c.deleteClusterServiceBroker(); err != nil {
		return errors.Wrap(err, "failed during removing ClusterServiceBroker")
	}
	return nil
}

func (c *cleaner) serviceBindingExist() (bool, error) {
	_, err := c.sc.ServiceBindings(c.namespace).Get(serviceBindingName, metav1.GetOptions{})
	if apiErr.IsNotFound(err) {
		klog.Infof("ServiceBinding %q not exist", serviceBindingName)
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c *cleaner) deleteServiceBinding() error {
	err := c.sc.ServiceBindings(c.namespace).Delete(serviceBindingName, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (c *cleaner) assertServiceBindingIsRemoved() error {
	return wait.Poll(waitInterval, timeoutInterval, func() (done bool, err error) {
		_, err = c.sc.ServiceBindings(c.namespace).Get(serviceBindingName, metav1.GetOptions{})
		if apiErr.IsNotFound(err) {
			return true, nil
		}
		if err != nil {
			return false, err
		}

		return false, nil
	})
}

func (c *cleaner) serviceInstanceExist() (bool, error) {
	_, err := c.sc.ServiceInstances(c.namespace).Get(serviceInstanceName, metav1.GetOptions{})
	if apiErr.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c *cleaner) removeServiceInstanceFinalizer() error {
	instance, err := c.sc.ServiceInstances(c.namespace).Get(serviceInstanceName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	toUpdate := instance.DeepCopy()
	toUpdate.Finalizers = nil

	_, err = c.sc.ServiceInstances(toUpdate.Namespace).Update(toUpdate)
	if err != nil {
		return err
	}

	return nil
}

func (c *cleaner) deleteServiceInstance() error {
	err := c.sc.ServiceInstances(c.namespace).Delete(serviceInstanceName, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (c *cleaner) assertServiceInstanceIsRemoved() error {
	return wait.Poll(waitInterval, timeoutInterval, func() (done bool, err error) {
		_, err = c.sc.ServiceInstances(c.namespace).Get(serviceInstanceName, metav1.GetOptions{})
		if apiErr.IsNotFound(err) {
			return true, nil
		}
		if err != nil {
			return false, err
		}

		return false, nil
	})
}

func (c *cleaner) deleteClusterServiceBroker() error {
	err := c.sc.ClusterServiceBrokers().Delete(clusterServiceBrokerName, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}
