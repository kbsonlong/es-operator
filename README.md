# es-operator
// TODO(user): Add simple overview of use/purpose

## Description
// TODO(user): An in-depth paragraph about your project and overview of use

## Getting Started
You’ll need a Kubernetes cluster to run against. You can use [KIND](https://sigs.k8s.io/kind) to get a local cluster for testing, or run against a remote cluster.
**Note:** Your controller will automatically use the current context in your kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).


```yaml
cat << EOF > kind-cluster.yaml
kind: Cluster
apiVersion: "kind.x-k8s.io/v1alpha4"
networking:
  podSubnet: "10.10.0.0/16"
  serviceSubnet: "10.11.0.0/16"
nodes:
  - role: control-plane
    image: registry.cn-hangzhou.aliyuncs.com/seam/node:v1.24.15
EOF
```


```bash
kind create cluster --name cluster --config=kind-cluster.yaml
```



### Running on the cluster
1. Install Instances of Custom Resources:

```sh
kubectl apply -f config/samples/
```

2. Build and push your image to the location specified by `IMG`:
	
```sh
make docker-build docker-push IMG=<some-registry>/es-operator:tag
```
	
3. Deploy the controller to the cluster with the image specified by `IMG`:

```sh
make deploy IMG=<some-registry>/es-operator:tag
```

### Uninstall CRDs
To delete the CRDs from the cluster:

```sh
make uninstall
```

### Undeploy controller
UnDeploy the controller to the cluster:

```sh
make undeploy
```

## Contributing
// TODO(user): Add detailed information on how you would like others to contribute to this project

### How it works
This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/) 
which provides a reconcile function responsible for synchronizing resources untile the desired state is reached on the cluster 



### Debug Code on local environment

使用 [telepresence](https://www.telepresence.io/docs/latest/quick-start/) 进行本地调试

1. Install telepresence into the cluster:

```bash
~ telepresence helm install
...
Traffic Manager installed successfully

~ kubectl get pod -n ambassador
NAME                               READY   STATUS    RESTARTS   AGE
traffic-manager-75bcdb8dc9-rhthl   1/1     Running   0          22m
```

2. Connect telepresence on local machine

```bash
~ telepresence connect
Connected to context kind-cluster (https://172.26.128.224:49669)

~ telepresence status
User Daemon: Running
  Version           : v2.13.1
  Executable        : /usr/local/bin/telepresence
  Install ID        : e69cbab8-e5b2-4e10-b6d4-1c023e45161e
  Status            : Connected
  Kubernetes server : https://172.26.128.224:49669
  Kubernetes context: kind-cluster
  Manager namespace : ambassador
  Intercepts        : 0 total
Root Daemon: Running
  Version    : v2.13.1
  Version    : v2.13.1
  DNS        :
    Remote IP       : 127.0.0.1
    Exclude suffixes: [.com .io .net .org .ru]
    Include suffixes: []
    Timeout         : 8s
  Also Proxy : (0 subnets)
  Never Proxy: (1 subnets)
    - 172.26.128.224/32
Ambassador Cloud:
  Status      : Logged out
Traffic Manager: Connected
  Version : v2.13.1
  Mode    : single-user
Intercept Spec: Not running
```

3. Test telepresence

```bash
➜  ~ kubectl get pod -n kube-system -l k8s-app=kube-dns -o wide
NAME                       READY   STATUS    RESTARTS   AGE   IP          NODE                    NOMINATED NODE   READINESS GATES
coredns-57575c5f89-ftr86   1/1     Running   0          26m   10.10.0.4   cluster-control-plane   <none>           <none>
coredns-57575c5f89-ppnhf   1/1     Running   0          26m   10.10.0.2   cluster-control-plane   <none>           <none>
➜  ~ ping 10.10.0.4
PING 10.10.0.4 (10.10.0.4): 56 data bytes
64 bytes from 10.10.0.4: icmp_seq=0 ttl=64 time=0.846 ms
64 bytes from 10.10.0.4: icmp_seq=1 ttl=64 time=0.750 ms
^C
--- 10.10.0.4 ping statistics ---
2 packets transmitted, 2 packets received, 0.0% packet loss
round-trip min/avg/max/stddev = 0.750/0.798/0.846/0.048 ms
➜  ~ telnet 10.10.0.4 53
Trying 10.10.0.4...
Connected to 10.10.0.4.
Escape character is '^]'.
^]
telnet> q
Connection closed.
➜  ~ nslookup kubernetes.default.svc.cluster.local 10.10.0.4
Server:		10.10.0.4
Address:	10.10.0.4#53

Name:	kubernetes.default.svc.cluster.local
Address: 10.11.0.1
```


### Test It Out
1. Install the CRDs into the cluster:

```sh
make install
```

2. Run your controller (this will run in the foreground, so switch to a new terminal if you want to leave it running):

```sh
make run
```

**NOTE:** You can also run this in one step by running: `make install run`

### Modifying the API definitions
If you are editing the API definitions, generate the manifests such as CRs or CRDs using:

```sh
make manifests
```

**NOTE:** Run `make --help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

