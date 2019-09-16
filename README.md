k3v - Virtual Kubernetes
========

**STATUS: Proof Of Concept**

k3v runs as a dedicated virtual Kubernetes control plane.  When pods are
launched they are launched in another cluster.  k3v uses the compute,
storage, and networking resources from a real Kubernetes cluster. This
allows one to take one physical Kubernetes cluster and chop it up into
smaller virtual clusters. Also, it is theoretically possible that one
virtual cluster spans multiple physical clusters.

## Quick Start

Please note, this is POC quality stuff. 

1. Download k3v binary from the releases page.
    ***NOTE: While Windows and macOS binaries are available they probably won't work.***
2. Run k3v pointing to a kubeconfig file of your physical cluster.
    ```sh
    ./k3v --kubeconfig someconfig.yaml
    ```
   If you kube config is in the standard `$HOME/.kube/config` then no
   argument is needed.
3. Once started k3v will create a local folder `./k3v-data` that has the
   virtual kubernetes state.  Also `./kubeconfig.yaml` will be created.
4. Use `./kubeconfig.yaml` to talk to k3v
    ```sh
   kubectl --kubeconfig ./kubeconfig.yaml --all-namespaces=true get all
    ```

## Benefits

The reasons for experimenting with virtual clusters are

1. Better security/multitenancy
2. Better separation of concerns between infra and custom controllers (operators) 
3. Ability to package complex k8s based applications

### Security

Virtual clusters can help create a better model for hard multi tenancy.

Multitenancy in Kubernetes is hard for multiple reasons.  First you must
trust the security of a container.  Meaning that one bad neighbor can't
attack another neighbor. Virtual clusters does not help with this concern.
To address this you need to trust containers or leverage another technology
like gvisor or katacontainers.  The second and more fundamental issue is
that the attack surface of Kubernetes is far to large for multitenancy.  Right
now the only way to accomplish multi tenancy is to not allow users to do the
vast majority of Kubernetes operations.  It is far too difficult to ensure that
all the various APIs will not expose some issue.  But limiting access means
an end user can't leverage a lot of functionality of Kubernetes, such as
operators.

Virtual clusters allow you to separate out the problem into two distinct layers.
First you need a secure and very limited physical cluster.  A virtual cluster only
requires very basic CRUD privileges on pods, services, endpoints, configmap,
secrets, and pvcs. It is far easier to secure this small set of APIs. Then
tenants can be given access to a virtual kubernetes instance and given full
cluster admin privileges.  They can do whatever they want with Kubernetes and
you can be confident they won't impact another neighbor.

### Separation of Concerns

Kubernetes is a great architecture that allows one to write things such as
operators or controllers to manage your infrastructure.  Unfortunately each
operator or controller you add to a cluster could have larger impacts on the
cluster.  Today, if you want an operator, really a cluster admin must install
that for you.  And more importantly, the cluster admin must have confidence
the operator won't do harm to the cluster.  This is not a scalable model.

With virtual clusters you can setup large, rather boring physical clusters. One
team can be responsible for the core uptime of this cluster.  Then each virtual
cluster can install operators or any random component as they wish, with a high
degree of confidence that if something goes wrong it only impacts this one
virtual cluster.

### Kubernetes Based App Packaging

If you write a complex Kubernetes based application (think istio or such) one
could package this application in a virtual cluster and deploy it in a cluster
much like a normal controller, but with no CRDs. No matter what madness that
application does (CRDs, finalizers, webhooks, etc) it is all nicely encapsulated.

## License
Copyright (c) 2019 [Rancher Labs, Inc.](http://rancher.com)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

[http://www.apache.org/licenses/LICENSE-2.0](http://www.apache.org/licenses/LICENSE-2.0)

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
