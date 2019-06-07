package service_broker

import (
	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	scClientset "github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	apiErr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog"
)

type tester struct {
	common
	c         scClientset.ServicecatalogV1beta1Interface
	namespace string
}

func newTester(cli ClientGetter, ns string) *tester {
	return &tester{
		c:         cli.ServiceCatalogClient().ServicecatalogV1beta1(),
		namespace: ns,
		common: common{
			sc:        cli.ServiceCatalogClient().ServicecatalogV1beta1(),
			namespace: ns,
		},
	}
}

func (t *tester) execute() error {
	klog.Info("Start test resources for ServiceBroker test")
	for _, fn := range []func() error{
		t.assertServiceBrokerIsReady,
		t.checkServiceClass,
		t.checkServicePlan,
		t.assertServiceInstanceIsReady,
		t.assertServiceBindingIsReady,
	} {
		err := fn()
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *tester) assertServiceBrokerIsReady() error {
	return wait.Poll(waitInterval, timeoutInterval, func() (done bool, err error) {
		broker, err := t.sc.ServiceBrokers(t.namespace).Get(serviceBrokerName, metav1.GetOptions{})
		if apiErr.IsNotFound(err) {
			klog.Infof("ServiceBroker %q not exist", serviceBrokerName)
			return false, nil
		}
		if err != nil {
			return false, err
		}

		condition := v1beta1.ServiceBrokerCondition{
			Type:    v1beta1.ServiceBrokerConditionReady,
			Status:  v1beta1.ConditionTrue,
			Message: successFetchedCatalogMessage,
		}
		for _, cond := range broker.Status.Conditions {
			if condition.Type == cond.Type && condition.Status == cond.Status && condition.Message == cond.Message {
				klog.Info("ServiceBroker is in ready state")
				return true, nil
			}
			klog.Infof("ServiceBroker is not ready, condition: Type: %q, Status: %q, Reason: %q", cond.Type, cond.Status, cond.Message)
		}

		return false, nil
	})
}
