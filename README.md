# cluster-api-inception

This is a Cluster API provider that creates all the components of a Kubernetes cluster .. without creating anything.

## Pre-requisites

- A Kubernetes cluster
- A "large" cup of tea
- Internet access

### Install Cluster-API

Installing Cluster-API is as simple as importing the associated `yaml` with the below command:

```
kubectl create -f https://github.com/kubernetes-sigs/cluster-api/releases/download/v0.2.6/cluster-api-components.yaml
```
We can verify that it is all installed by looking at the `crd`s or the `clusterRole` bindings. 

### Install Cluster-API-Inception

As with the cluster-API provider we can install the inception provider by importing the relevant `yaml`

Install the CRDs

```
kubectl create -f https://github.com/thebsdbox/cluster-api-inception/raw/master/all-the-yaml/crd.yaml
```

Install the controller

```
kubectl create -f https://github.com/thebsdbox/cluster-api-inception/raw/master/all-the-yaml/controller.yaml
```

## Deploy a cluster

Inside this repository is a fake big-ish cluster, we can clone/download this repository locally and then navigate to the `all-the-yaml/big` directory. We can then apply all the cluster/machineDeployments with the following command:

```
~/cluster-api-inception/all-the-yaml/big/ $ kubectl apply -f .
```

We can now inspect and scale this cluster:

```
$ k get clusters
NAME                    PHASE
cluster-new-york        provisioned
cluster-portland        provisioned
cluster-san-francisco   provisioned
cluster-seattle         provisioned
cluster-sheffield       provisioned
$ k get machinedeployments
NAME                       AGE
deployment-new-york        14m
deployment-portland        14m
deployment-san-francisco   14m
deployment-seattle         14m
deployment-sheffield       14m
$ k get machines | grep seattle
deployment-seattle-5d86486bff-4nfwq         iProvider    provisioned
deployment-seattle-5d86486bff-6bxl4         iProvider    provisioned
deployment-seattle-5d86486bff-6jh9m         iProvider    provisioned
deployment-seattle-5d86486bff-6qd6k         iProvider    provisioned
deployment-seattle-5d86486bff-ckn2v         iProvider    provisioned
$ k scale --replicas=25 machinedeployment/deployment-seattle 
machinedeployment.cluster.x-k8s.io/deployment-seattle scaled
$ k get machines | grep seattle 
deployment-seattle-5d86486bff-4nfwq         iProvider    provisioned
deployment-seattle-5d86486bff-6bxl4         iProvider    provisioned
deployment-seattle-5d86486bff-6jbwp                      provisioning
deployment-seattle-5d86486bff-6jh9m         iProvider    provisioned
deployment-seattle-5d86486bff-6qd6k         iProvider    provisioned
deployment-seattle-5d86486bff-8lq9h                      provisioning
deployment-seattle-5d86486bff-bw9dr                      provisioning
deployment-seattle-5d86486bff-ckn2v         iProvider    provisioned
deployment-seattle-5d86486bff-ff4bg                      provisioning
deployment-seattle-5d86486bff-h6jtn         iProvider    provisioned
deployment-seattle-5d86486bff-h6px4                      provisioning
deployment-seattle-5d86486bff-jfw6z                      provisioning
deployment-seattle-5d86486bff-mk269                      provisioning
deployment-seattle-5d86486bff-mtpkm         iProvider    provisioned
deployment-seattle-5d86486bff-nrpwd         iProvider    provisioned
deployment-seattle-5d86486bff-pkhm2         iProvider    provisioned
deployment-seattle-5d86486bff-r74zh                      provisioning
deployment-seattle-5d86486bff-rb7h7                      provisioning
deployment-seattle-5d86486bff-tcbsw                      provisioning
deployment-seattle-5d86486bff-tzqh6                      provisioning
deployment-seattle-5d86486bff-vq9t9                      provisioning
deployment-seattle-5d86486bff-x8wsd                      provisioning
deployment-seattle-5d86486bff-xxdpv                      provisioning
deployment-seattle-5d86486bff-zf92j                      provisioning
deployment-seattle-5d86486bff-znzkg                      provisioning
```

## Cluster-API bug

If you delete a cluster whilst the machines are still being deleted, they may be stuck in a `deleting` state. In order to rectify this.. re-create the associated cluster and wait until machines have been cleaned, then remove the cluster. Details => https://github.com/kubernetes-sigs/cluster-api/issues/1643
