## Tutorial: Lobster on minikube

### You can launch `lobster` in `minikube`

- What you need: 
  - `minikube`: https://minikube.sigs.k8s.io/
  - `kubectl`: https://kubernetes.io/docs/tasks/tools/
  - `helm`: https://helm.sh/
- There are various container runtimes in minikube. \
  In this tutorial, you can run Lobster according to the runtime environment.

### Build images in the local registry

> Provision of public images is under legal review. \
  So I recommend using a local registry for now.

#### 1. Setup registry in minikube

See https://minikube.sigs.k8s.io/docs/handbook/registry/

#### 2. Push registry

```bash
$ make REGISTRY=localhost:5000 push-all
```

##### Check results

```bash
$ docker image ls | grep lobster                                                                                                  

localhost:5000/lobster-exporter                                     vX.X.X      ...
localhost:5000/lobster-syncer                                       vX.X.X      ...
localhost:5000/lobster-global-query                                 vX.X.X      ...
localhost:5000/lobster-query                                        vX.X.X      ...
localhost:5000/lobster-store                                        vX.X.X      ...
```

### Run the Lobster

#### 1. Launch minkube

##### docker

```bash
$ minikube start --container-runtime=docker
```

##### containerd

```bash
$ minikube start --container-runtime=containerd
```

#### 2. Install Lobster set

- Install lobster using the `helm` command
- It is necessary to distinguish between options depending on the container runtime used

##### docker
- If the log is in [json format](https://docs.docker.com/engine/logging/drivers/json-file/), add `loglineFormat: json`
- Mount the docker root directory (`/var/lib/docker`) where the actual logs are located

```bash
$ helm upgrade --install lobster_cluster ./deploy -f ./deploy/values/public/lobster-cluster_basic.yaml \
--set registry=localhost:5000 \
--set loglineFormat=json
```

##### containerd
- add `loglineFormat: text` or not(default)

```bash
$ helm upgrade --install lobster_cluster ./deploy -f ./deploy/values/public/lobster-cluster_basic.yaml \
--set registry=localhost:5000 

or

$ helm upgrade --install lobster_cluster ./deploy -f ./deploy/values/public/lobster-cluster_basic.yaml \
--set registry=localhost:5000 \
--set loglineFormat=text 
```

##### Check results

```bash
$ kubectl get pod 
NAME                             READY   STATUS    RESTARTS   AGE
lobster-query-0-7c486fb4-fhqkx   1/1     Running   0          51s
lobster-store-n2cs5              1/1     Running   0          51s
loggen-5676c84f9b-bn7q9          2/2     Running   0          51s
```

#### 3. Access to Lobster web page

- Perform port-forward to obtain an externally accessible address
- You can check the log by accessing http://127.0.0.1:8080
- You can search logs produced in [Loggen](./design/loggen.md) based on Kubernetes object

```bash
$ kubectl port-forward $(kubectl get pod -l app=lobster-query --no-headers | awk '{print$1}') 8080:80

Forwarding from 127.0.0.1:8080 -> 80
Forwarding from [::1]:8080 -> 80
```

[](./images/tutorial.gif)

For more features, please refer to the [deployment guide](deployment.md).