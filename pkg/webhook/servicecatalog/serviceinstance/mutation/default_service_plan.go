package mutation

import (
	"net/http"
	"fmt"
	"errors"
	"context"

	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	sc "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"github.com/kubernetes-incubator/service-catalog/pkg/webhookutil"
)

type DefaultServicePlan struct {
	Client   client.Client
	Ctx      context.Context
	Instance *sc.ServiceInstance
	log      *webhookutil.TracedLogger
}

func (d *DefaultServicePlan) HandleDefaultPlan() *MutateError {
	if d.Instance.Spec.ClusterServicePlanSpecified() || d.Instance.Spec.ServicePlanSpecified() {
		return nil
	}

	if d.Instance.Spec.ClusterServiceClassSpecified() {
		return d.handleDefaultClusterServicePlan()
	} else if d.Instance.Spec.ServiceClassSpecified() {
		return d.handleDefaultServicePlan()
	}

	return NewMutateError("class not specified on ServiceInstance, cannot choose default plan", http.StatusInternalServerError)
}

func (d *DefaultServicePlan) handleDefaultClusterServicePlan() *MutateError {
	sc, err := d.getClusterServiceClassByPlanReference()
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return NewMutateError(err.Error(), http.StatusForbidden)
		}
		msg := fmt.Sprintf("ClusterServiceClass %c does not exist, can not figure out the default ClusterServicePlan.",
			d.Instance.Spec.PlanReference)
		d.log.Info(msg)
		return NewMutateError(msg, http.StatusForbidden)
	}

	// find all the service plans that belong to the service class

	// Need to be careful here. Is it possible to have only one
	// ClusterServicePlan available while others are still in progress?
	// Not currently. Creation of all ClusterServicePlans before creating
	// the ClusterServiceClass ensures that this will work correctly. If
	// the order changes, we will need to rethink the
	// implementation of this controller.
	plans, err := d.getClusterServicePlansByClusterServiceClassName(sc.Name)
	if err != nil {
		msg := fmt.Sprintf("Error listing ClusterServicePlans for ClusterServiceClass (K8S: %v ExternalName: %v) - retry and specify desired ClusterServicePlan", sc.Name, sc.Spec.ExternalName)
		d.log.Infof(`ServiceInstance "%s/%s": %s`, d.Instance.Namespace, d.Instance.Name, msg)
		return NewMutateError(msg, http.StatusForbidden)
	}

	// check if there were any service plans
	// TODO: in combination with not allowing classes with no plans, this should be impossible
	if len(plans) == 0 {
		msg := fmt.Sprintf("no ClusterServicePlans found at all for ClusterServiceClass %q", sc.Spec.ExternalName)
		d.log.Infof(`ServiceInstance "%s/%s": %s`, d.Instance.Namespace, d.Instance.Name, msg)
		return NewMutateError(msg, http.StatusForbidden)
	}

	// check if more than one service plan was found and error
	if len(plans) > 1 {
		msg := fmt.Sprintf("ClusterServiceClass (K8S: %v ExternalName: %v) has more than one plan, PlanName must be specified", sc.Name, sc.Spec.ExternalName)
		d.log.Infof(`ServiceInstance "%s/%s": %s`, d.Instance.Namespace, d.Instance.Name, msg)
		return NewMutateError(msg, http.StatusForbidden)
	}
	// otherwise, by default, pick the only plan that exists for the service class

	p := plans[0]
	d.log.Infof(`ServiceInstance "%s/%s": Using default plan %q (K8S: %q) for Service Class %q`,
		d.Instance.Namespace, d.Instance.Name, p.Spec.ExternalName, p.Name, sc.Spec.ExternalName)
	if d.Instance.Spec.ClusterServiceClassExternalName != "" {
		d.Instance.Spec.ClusterServicePlanExternalName = p.Spec.ExternalName
	} else if d.Instance.Spec.ClusterServiceClassExternalID != "" {
		d.Instance.Spec.ClusterServicePlanExternalID = p.Spec.ExternalID
	} else {
		d.Instance.Spec.ClusterServicePlanName = p.Name
	}

	return &MutateError{}
}

func (d *DefaultServicePlan) handleDefaultServicePlan() *MutateError {
	sc, err := d.getServiceClassByPlanReference()
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return NewMutateError(err.Error(), http.StatusForbidden)
		}
		msg := fmt.Sprintf("ServiceClass %c does not exist, can not figure out the default ServicePlan.",
			d.Instance.Spec.PlanReference)
		d.log.Info(msg)
		return NewMutateError(msg, http.StatusForbidden)
	}
	// find all the service plans that belong to the service class

	// Need to be careful here. Is it possible to have only one
	// ServicePlan available while others are still in progress?
	// Not currently. Creation of all ServicePlans before creating
	// the ServiceClass ensures that this will work correctly. If
	// the order changes, we will need to rethink the
	// implementation of this controller.
	plans, err := d.getServicePlansByServiceClassName(sc.Name)
	if err != nil {
		msg := fmt.Sprintf("Error listing ServicePlans for ServiceClass (K8S: %v ExternalName: %v) - retry and specify desired ServicePlan", sc.Name, sc.Spec.ExternalName)
		d.log.Infof(`ServiceInstance "%s/%s": %s`, d.Instance.Namespace, d.Instance.Name, msg)
		return NewMutateError(msg, http.StatusForbidden)
	}

	// check if there were any service plans
	// TODO: in combination with not allowing classes with no plans, this should be impossible
	if len(plans) == 0 {
		msg := fmt.Sprintf("no ServicePlans found at all for ServiceClass %q", sc.Spec.ExternalName)
		d.log.Infof(`ServiceInstance "%s/%s": %s`, d.Instance.Namespace, d.Instance.Name, msg)
		return NewMutateError(msg, http.StatusForbidden)
	}

	// check if more than one service plan was found and error
	if len(plans) > 1 {
		msg := fmt.Sprintf("ServiceClass (K8S: %v ExternalName: %v) has more than one plan, PlanName must be specified", sc.Name, sc.Spec.ExternalName)
		d.log.Infof(`ServiceInstance "%s/%s": %s`, d.Instance.Namespace, d.Instance.Name, msg)
		return NewMutateError(msg, http.StatusForbidden)
	}
	// otherwise, by default, pick the only plan that exists for the service class

	p := plans[0]
	d.log.Infof(`ServiceInstance "%s/%s": Using default plan %q (K8S: %q) for Service Class %q`,
		d.Instance.Namespace, d.Instance.Name, p.Spec.ExternalName, p.Name, sc.Spec.ExternalName)
	if d.Instance.Spec.ServiceClassExternalName != "" {
		d.Instance.Spec.ServicePlanExternalName = p.Spec.ExternalName
	} else if d.Instance.Spec.ServiceClassExternalID != "" {
		d.Instance.Spec.ServicePlanExternalID = p.Spec.ExternalID
	} else {
		d.Instance.Spec.ServicePlanName = p.Name
	}

	return &MutateError{}
}

func (d *DefaultServicePlan) getClusterServiceClassByPlanReference() (*sc.ClusterServiceClass, error) {
	if d.Instance.Spec.PlanReference.ClusterServiceClassName != "" {
		return d.getClusterServiceClassByK8SName()
	}

	return d.getClusterServiceClassByField()
}

func (d *DefaultServicePlan) getServiceClassByPlanReference() (*sc.ServiceClass, error) {
	if d.Instance.Spec.PlanReference.ServiceClassName != "" {
		return d.getServiceClassByK8SName()
	}

	return d.getServiceClassByField()
}

func (d *DefaultServicePlan) getClusterServiceClassByK8SName() (*sc.ClusterServiceClass, error) {
	d.log.Infof("Fetching ClusterServiceClass by k8s name %q", d.Instance.Spec.PlanReference.ClusterServiceClassName)
	csc := &sc.ClusterServiceClass{}
	err := d.Client.Get(d.Ctx, client.ObjectKey{Name: d.Instance.Spec.PlanReference.ClusterServiceClassName}, csc)
	return csc, err
}

func (d *DefaultServicePlan) getServiceClassByK8SName() (*sc.ServiceClass, error) {
	d.log.Infof("Fetching ServiceClass by k8s name %q", d.Instance.Spec.PlanReference.ServiceClassName)
	sc := &sc.ServiceClass{}
	err := d.Client.Get(d.Ctx, client.ObjectKey{Name: d.Instance.Spec.PlanReference.ServiceClassName, Namespace: d.Instance.Namespace}, sc)

	return sc, err
}

func (d *DefaultServicePlan) getClusterServiceClassByField() (*sc.ClusterServiceClass, error) {
	ref := d.Instance.Spec.PlanReference

	filterLabel := ref.GetClusterServiceClassFilterLabelName()
	filterValue := ref.GetSpecifiedClusterServiceClass()

	d.log.Infof("Fetching ClusterServiceClass filtered by %q = %q", filterLabel, filterValue)

	serviceClassesList := &sc.ClusterServiceClassList{}
	err := d.Client.List(d.Ctx, serviceClassesList, client.MatchingLabels(map[string]string{
		filterLabel: filterValue,
	}))
	if err != nil {
		d.log.Infof("Listing ClusterServiceClasses failed: %q", err)
		return nil, err
	}
	if len(serviceClassesList.Items) == 1 {
		d.log.Infof("Found single ClusterServiceClass as %+v", serviceClassesList.Items[0])
		return &serviceClassesList.Items[0], nil
	}
	msg := fmt.Sprintf("Could not find a single ClusterServiceClass with %q = %q, found %v", filterLabel, filterValue, len(serviceClassesList.Items))
	d.log.Info(msg)
	return nil, errors.New(fmt.Sprintf("could not find a single ClusterServiceClass with %q = %q, found %v", filterLabel, filterValue, len(serviceClassesList.Items)))
}

func (d *DefaultServicePlan) getServiceClassByField() (*sc.ServiceClass, error) {
	ref := d.Instance.Spec.PlanReference

	filterLabel := ref.GetServiceClassFilterLabelName()
	filterValue := ref.GetSpecifiedServiceClass()

	d.log.Infof("Fetching ServiceClass filtered by %q = %q", filterLabel, filterValue)

	serviceClassesList := &sc.ServiceClassList{}
	err := d.Client.List(d.Ctx, serviceClassesList, client.MatchingLabels(map[string]string{
		filterLabel: filterValue,
	}))
	if err != nil {
		d.log.Infof("Listing ServiceClasses failed: %q", err)
		return nil, err
	}
	if len(serviceClassesList.Items) == 1 {
		d.log.Infof("Found single ServiceClass as %+v", serviceClassesList.Items[0])
		return &serviceClassesList.Items[0], nil
	}
	msg := fmt.Sprintf("Could not find a single ServiceClass with %q = %q, found %v", filterLabel, filterValue, len(serviceClassesList.Items))
	d.log.Info(msg)
	return nil, errors.New(fmt.Sprintf("could not find a single ServiceClass with %q = %q, found %v", filterLabel, filterValue, len(serviceClassesList.Items)))
}

// getClusterServicePlansByClusterServiceClassName() returns a list of
// ServicePlans for the specified service class name
func (d *DefaultServicePlan) getClusterServicePlansByClusterServiceClassName(scName string) ([]sc.ClusterServicePlan, error) {
	d.log.Infof("Fetching ClusterServicePlans by class name %q", scName)

	servicePlansList := &sc.ClusterServicePlanList{}
	err := d.Client.List(d.Ctx, servicePlansList, client.MatchingLabels(map[string]string{
		sc.GroupName + "/spec.clusterServiceClassRef.name": scName,
	}))
	if err != nil {
		d.log.Infof("Listing ClusterServicePlans failed: %q", err)
		return nil, err
	}

	d.log.Infof("ClusterServicePlans fetched by filtering classname: %+v", servicePlansList.Items)
	r := servicePlansList.Items
	return r, err
}

// getServicePlansByServiceClassName() returns a list of
// ServicePlans for the specified service class name
func (d *DefaultServicePlan) getServicePlansByServiceClassName(scName string) ([]sc.ServicePlan, error) {
	d.log.Infof("Fetching ServicePlans by class name %q", scName)

	servicePlansList := &sc.ServicePlanList{}
	err := d.Client.List(d.Ctx, servicePlansList, client.MatchingLabels(map[string]string{
		sc.GroupName + "/spec.serviceClassRef.name": scName,
	}))
	if err != nil {
		d.log.Infof("Listing ServicePlans failed: %q", err)
		return nil, err
	}
	d.log.Infof("ServicePlans fetched by filtering classname: %+v", servicePlansList.Items)
	r := servicePlansList.Items
	return r, err
}


func (d *DefaultServicePlan) InjectClient(c client.Client) error {
	d.Client = c
	return nil
}