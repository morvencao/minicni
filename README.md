# minicni

A simple CNI plugin implementation for kubernetes written in golang.

## TL;DR

Read the following articles about container and kubernetes network:

- [容器网络(一)](https://morven.life/posts/networking-4-docker-sigle-host/)
- [容器网络(二)](https://morven.life/posts/networking-5-docker-multi-hosts/)
- [浅聊 Kubernetes 网络模型](https://morven.life/posts/networking-6-k8s-summary/)
- [Container Network Interface Specification](https://github.com/containernetworking/cni/blob/master/SPEC.md)

This repo is responsible for implementing a Kubernetes overlay network, as well as for allocating and configuring network interfaces in pods. With minicni plugin installed into a kubernetes cluster, it should be able to achieve the following targets:

- All the podd can communicate with each other directly without NAT.
- All the nodes can communicate with all pods (and vice versa) without NAT.
- The IP that a pod sees itself as is the same IP that others see it as.

## Prerequisites

A running kubernetes cluster without any CNI plugins installed. There are many kubernetes installer tools, but [kubeadm](https://kubernetes.io/docs/reference/setup-tools/kubeadm/) is the most flexible one as it allows the use of your own network plug-in. Read the official doc for how to [Creating a cluster with kubeadm](https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/create-cluster-kubeadm/)

## Build and Test

1. build the minicni binary:

```
make build
```

2. build and push the minicni installer image:

```
IMAGE_REPO=<YOUR-DOCKER-REPO> IMAGE_NAME = install-minicni make image
```

> Note: Login your docker registry before pushing the minicni installer image.

3. Deploy the minicni installer into your kubernetes cluster:

```
kubectl apply -f deployments/manifests/minicni.yaml
```

4. Verify the minicni is installed successfully:

```
# kubectl -n kube-system get pod -l app=minicni
NAME                 READY   STATUS    RESTARTS   AGE
minicni-node-7dmsw   1/1     Running   0          38m
minicni-node-87c45   1/1     Running   0          38m
```

5. Deploy the test pods with networking debug tools into your kubernetes cluster:

```
kubectl apply -f tests/test-pods.yaml
```

> Note: Make sure to label the master and worker node so that the testing pods can be scheduled to correct node.

6. Verify the networking connections:

- pod to host node
- pod to other nodes
- pod to pod in the same node
- pod to pod across nodes

## Known issues:

1. By default pod-to-pod traffic is drop by the linux kernel because linux treats interfaces in non-root network namespaces as if they were external, see discussion [here](https://serverfault.com/questions/162366/iptables-bridge-and-forward-chain) To workaround this, we need to manually to add the following iptables rules in each cluster node:

```
iptables -t filter -A FORWARD -s <POD_CIDR> -j ACCEPT
iptables -t filter -A FORWARD -d <POD_CIDR> -j ACCEPT
```

2. For pod-to-pod communications across nodes, we need to add host gateway routes just like what [Calico](https://docs.projectcalico.org/networking/openstack/host-routes) does, the feature will be added in the future. For now, we have to manually to add the following route rules in each node:

```
ip route add 172.18.1.0/24 via 10.11.97.173 dev ens4 # run on master 
ip route add 172.18.0.0/24 via 10.11.97.64 dev ens4 # run on worker
```

> Note: In the command above, we have one master and one worker node, `172.18.1.0/24` is subnet for the worker node, `10.11.97.173` is the IPv4 address of the worker node; `172.18.0.0/24` is the subnet for the master node, `10.11.97.64` is the IPv4 address of the master node.

