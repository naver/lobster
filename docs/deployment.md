## Deployment

- Both helm v2 and v3 are available

### Organize lobster components according to your requirements

Components:
- [Cluster basic](../deploy/values/public/lobster-cluster_basic.yaml)
  - [lobster-query](./design/lobster_query.md), [lobster-store](./design/lobster_store.md)
- [Cluster with log sink](../deploy/values/public/lobster-cluster_logsink-extension.yaml)
  -  [lobster-query](./design/lobster_query.md), [lobster-store](./design/lobster_store.md), [lobster-syncer](./design/log_sink.md)
  - [Log sink operator](../deploy/values/public/lobster-operator.yaml)
- [Global querier](../deploy/values/public/lobster-global-query.yaml)
- Optional: [loggen](./design/loggen.md)(live test)

Combinations:

- Support for log `APIs` and `web` operations in a `single cluster`
  - [Cluster basic](../deploy/values/public/lobster-cluster_basic.yaml)
- Support for log `APIs`, `web` and `metric/export` operations in a `single cluster`
  - [Cluster with log sink](../deploy/values/public/lobster-cluster_logsink-extension.yaml) + [Log sink operator](../deploy/values/public/lobster-operator.yaml)
- Support for log `APIs`, `web` and `metric/export` operations across `multiple clusters`
  - Main cluster: [Cluster with log sink](../deploy/values/public/lobster-cluster_logsink-extension.yaml) + [Log sink operator](../deploy/values/public/lobster-operator.yaml)
  - Other clusters: [Cluster with log sink](../deploy/values/public/lobster-cluster_logsink-extension.yaml)
- Support for log `APIs` and `web` operations on `one endpoint` across `multiple clusters`
  - Main cluster: [Cluster basic](../deploy/values/public/lobster-cluster_basic.yaml) + [Global querier](../deploy/values/public/lobster-global-query.yaml)
  - Other clusters: [Cluster basic](../deploy/values/public/lobster-cluster_basic.yaml)
- Support for log `APIs`, `web` and `metric/export` operations on one endpoint across `multiple clusters`
  - Main cluster: [Cluster with log sink](../deploy/values/public/lobster-cluster_logsink-extension.yaml) + [Global querier](../deploy/values/public/lobster-global-query.yaml) + [Log sink operator](../deploy/values/public/lobster-operator.yaml)
  - Other clusters: [Cluster with log sink](../deploy/values/public/lobster-cluster_logsink-extension.yaml)


### Deploying with helm charts

- The container log format may have different default settings depending on the container runtime
  - You can check it out in the [Tutorial: Lobster on minikube](tutorial.md)
- Depending on the container runtime(according to your environment), you may need to add the helm arguments below.
  - `--set loglineFormat=text` with continerd (default)
  - `--set loglineFormat=json` with docker

#### Cluster basic

```bash
helm upgrade --install --debug lobster_cluster -f ./deploy/values/public/lobster-cluster_basic.yaml 
```

#### Cluster with log sink
```bash
helm upgrade --install --debug lobster_cluster -f ./deploy/values/public/lobster-cluster_logsink-extension.yaml 
helm upgrade --install --debug lobster_operator -f ./deploy/values/public/lobster-operator.yaml 
```

#### Global querier

```bash
helm upgrade --install --debug lobster_global_query -f ./deploy/values/public/lobster-global-query.yaml
```
