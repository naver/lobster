Supports installing Lobster components using Helm charts. \

A simple set of logs for collecting and querying logs within a Kubernetes cluster is defined and \ 
an extended version installation that supports log metrics and export. \

The object to be installed is defined in the value file.\
Please replace the default value of each field with the value needed in your environment.

## helm chart
 
Change the `namespace` and `value` to suit your installation purpose.

```bash
helm upgrade --namespace {namespace} --install lobster ./deploy -f ./deploy/public/{value}.yaml 
```

## helm values

### lobster-cluster_basic

- Install Lobster components that support log collection and query serving

### lobster-cluster_logsink-extension

- Install Lobster components that support log collection, query serving and log sink(export/metric)
- To synchronize log sink rules, enter the address `lobster-operator` \
  The `lobster-syncer` requests `LobsterSink` rules via the `lobster-operator`
  - Path: `syncer.options.ruleStore`
  - Input: `{address}` (e.g. `lobster-operator:80`)

### lobster-global-query

- Install Lobster-global-query to query lobster installed in multiple clusters
- Please enter the address of the `lobster-query` service of the cluster where `lobster-cluster` is installed
  - Path: `global_query.options.lobsterQueries`
  - Input: `{cluster name}|{address}` (e.g. `local|lobster-query:80`)

### lobster-operator

- Install Lobster-operator to define log sinks(export/metric)
- Please refer to the [log sink docs](../../../docs/design/log_sink.md) for more details
