## Deployment

- Both helm v2 and v3 are available

### Organize lobster components according to your requirements

Components:
- [Cluster basic](../deploy/values/public/lobster-cluster_basic.yaml)
  - [lobster-query](./design/lobster_query.md), [lobster-store](./design/lobster_store.md)
- [Cluster with log sink](../deploy/values/public/lobster-cluster_logsink-extension.yaml)
  -  [lobster-query](./design/lobster_query.md), [lobster-store](./design/lobster_store.md), [lobster-syncer](./design/log_sink.md)
- [Global querier](../deploy/values/public/lobster-global-query.yaml)
- [Log sink operator](../deploy/values/public/lobster-operator.yaml)
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


#### Cluster basic

```bash
helm upgrade --install --debug -f ./deploy/values/public/lobster-cluster_basic.yaml 
```

#### Cluster with log sink
```bash
helm upgrade --install --debug -f ./deploy/values/public/lobster-cluster_logsink-extension.yaml 
```

#### Global querier

```bash
helm upgrade --install --debug -f ./deploy/values/public/lobster-global-query.yaml
```

#### Log sink operator

```bash
helm upgrade --install --debug -f ./deploy/values/public/lobster-operator.yaml 
```