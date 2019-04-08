# Service Catalog going towards to CRDs

 In Service Catalog we want to replace the Aggregated API with the CRD solution. This document specifies the concerns and possible solutions for supporting Service Catalog with an Aggregated API and the new Custom Resource Definitions (CRDs) approach.

## Support both Aggregated API and CRDs in the same code-base

Below you will find our concerns about having a single executable binary that supports both the Aggregated API and CRDs in the same code-base and changing the behavior by using a feature gate. 

### Business concerns

Adding CRDs via feature flag directly in Helm Chart can be misleading for the client. 
The CRDs it’s not a feature. It is just a new implementation for already existing feature. Leaving configuration [`apiserver.storage.type`: crds/etcd](https://github.com/kubernetes-incubator/service-catalog/blob/master/charts/catalog/values.yaml#L61-L62) is more like extending but not deprecating it. We do not want to make that customer think that **CRD** or **etcd** is an option as storage. Customers should know that the etcd is deprecated and they should consider switching for CRDs approach as soon as possible.  

Additionally, supporting bugs fixing and adding new features both in Aggregated API and CRDs could slow-down the development process.    

Another problem is to have consistent and up-to-date documentation. When we will support both approaches then we need to maintain documentation for them too. One more time it could be misleading for customers if we will show them that we are going in two directions at the same time.    

### Technical concerns

We need to clearly state that in Service Catalog the Aggregated API and the CRDs are not only about the underlying storage backend. Around those approaches, we have business logic. Because of that, we will end up with a lot of `if` statements in:
- controller reconcile process 

  |                                   | Aggregated API                      | CRDs                                                                                   |
  |-----------------------------------|-------------------------------------|----------------------------------------------------------------------------------------|
  | Queries                           | uses FieldSelector                  | use LabelSelector, cause the CRD does not support queries via Fields                   |
  | Removing finalizes                | in `UpdateStatus` method            | in `Update` method                                                                     |
  | ServiceInstance references fields | via custom `reference` sub-resource | does not support the generic sub-resources so, setting those fields that need to be changed |

- `svcat` CLI 

  |         | Aggregated API     | CRDs              |
  |---------|--------------------|-------------------|
  | Queries | uses FieldSelector | use LabelSelector |
  
- unit tests coverage - we had to also adjust unit tests because we have a slightly different approach, as you can see above. If we want to have one code base then we need to have doubled tests or support both flows in each test.
- validation of incoming Custom Resources (CR)

  | Aggregated API                                        | CRDs              |
  |-------------------------------------------------------|-------------------|
  | `ValidateUpdate` methods and some validation in plugins | Unified and realized only via **ValidatingWebhookConfiguration** |

- defaulting fields of incoming Custom Resources (CR) 

  | Aggregated API                                                      | CRDs                                                           |
  |---------------------------------------------------------------------|----------------------------------------------------------------|
  | `PrepareForUpdate` methods and defaults schemas via `defaulter-gen` | Unified and realised only via **MutatingWebhookConfiguration** |

- defining services in Helm Chart - different RBACs, secrets, deployments, services, etc.

- underlying Kubernetes different version constraint

> **NOTE:** As you can see above the Aggregated API use the validation and mutation in a different way. Sometimes same logic is split across different pkgs. In CRDs we unified that and copied it directly to the webhook domain. If we want to support both the API and CRD concept then we need to invest some time to extract it to some common libs. Then in both places use it with an overhead of adjusting the generic interfaces to a custom one. Additionally, after removing the API Server, we need to migrate them back to the domain as having it extracted will only mess the code and additional abstraction layer can confuse other developers.  
 
> **NOTE:** The described differences are those that we see from the short walk-through. Finally, there may be more differences - especially if we will support adding features both for api-server and CRDs.
 
**From the technical point of view having Aggregated API and CRDs in the same code-base it's not a feature gate only in the `service-catalog` binary but in the whole eco-system.** 

## Alternative Solution

Before merging the CRDs solution into the master, create a branch `release-0.1` with the last release of Service Catalog with Aggregated API server. 
On the `release-0.1` branch we are still doing the **bug fixing** for this version of Service Catalog but features are not introduced. When some bug will be found then we can fix it and still in an easy way create a release with version `0.1.x` 

In the master branch, we have Service Catalog with the CRD approach. New releases are created with the `0.2.x` version. On master branch, we are doing both the bug fixing and adding new features.

Such strategy will help us to get new customers cause we will show that the CRD is our new direction on which we want to focus. 
 
On the other hand, existing Service Catalog customers will see that the old solution is deprecated and the new features will be available only in the new approach. Thanks to that they will consider updating their system to the newest version as soon as possible cause the goal is to use the CRD solution as Kubernetes community recommends that.

We can set a reasonable time for supporting bug fixing for the Service Catalog with Aggregated API, e.g. 1 year.

With such solution, problems described in the previous section are mitigated.  
     
## Migration Support

If we want to convince existing customers to use the new Service Catalog version then we need to create a migration guide/scripts. The migration needs to be as simple as possible.

The migration will be simpler when we will hardcode for a given functionality in Aggregated API. 

#### Details

The migration logic can be placed in the Service Catalog Helm Chart. 
When someone wants to upgrade existing Service Catalog to the newest one, then he can execute `helm upgrade {release-name} svc-cat/catalog`

Raw scenario:
- **pre-upgrade hook**
  - replace the api-server deployment with the api-server that has a read-only mode
  - backup the Service Catalog resources to the persistent volume
- **upgrade**
  - removes the api-server
  - remove the etcd storage
  - adjust secrets, RBAC etc.
  - upgrade controller manager
  - install webhook server
- **post-upgrade**
  - scale down the controller manager to 0  
  - restore the Service Catalog resources - Spec and Status (status is important cause we do not want to trigger provisioning for already process items)
  - scale up controller manager


## Sum up

We need to be sure that customer will exactly know in which direction we want to go. Supporting only bug fixing for api-server could simplify adding new features while we will still provide support for older clients. The migration process will also be easier when we will hardcode for a given functionality in Aggregated API. For new contributors, it will be much easier to get familiar with the code base where only one approach is used. 

From our perspective having both approaches at the same time and supporting bug fixing and feature for both of them causing only that we are postponing the whole process. Sooner or later, we will have to take all these steps and back to the discussion. The only problem with doing that later can be that it will be harder (technical debt, confused customers, etc.)
