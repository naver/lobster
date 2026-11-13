Supports installing Lobster components using Helm charts. \

A simple set of logs for collecting and querying logs within a Kubernetes cluster is defined and \ 
an extended version installation that supports log metrics and export. \

The object to be installed is defined in the value file.\
Please replace the default value of each field with the value needed in your environment.

## helm chart
 
Change the `namespace` and `value` to suit your installation purpose.

```bash
helm upgrade --namespace {namespace} --install lobster ./deploy -f ./deploy/values/public/{value}.yaml 
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

## Auth token

`lobster-operator` can require an auth token on `GET /sync`, which `lobster-syncer` calls to get the log sink rules. Other endpoints are not checked.

Syncers in several clusters can sync from one operator, so its address may be open to other clusters. The token restricts who can call `GET /sync`.

- The syncer and the operator are given the same token. The chart keeps it in a Secret of each, and the syncer reads it on every request
- Without a token, the operator accepts any request

Use a random value without spaces as the token, for example:

```bash
openssl rand -hex 32
```

Keep the token in your deployment system and pass it when deploying the clusters and the operator:

```bash
helm upgrade --install lobster ./deploy -f {cluster values} --set syncer.options.authToken="$AUTH_TOKEN"
helm upgrade --install lobster-operator ./deploy -f {operator values} --set operator.options.authToken="$AUTH_TOKEN"
```

Set `syncer.options.authToken` in every cluster before `operator.options.authToken`, so that every syncer sends the token before the operator checks it.

- To rotate the token, give the operator both tokens separated by a space (`--set operator.options.authToken="$OLD_TOKEN $NEW_TOKEN"`), deploy every cluster with the new token, and then give the operator only the new one. A change to the operator's token restarts the operator
- A syncer that the operator refuses serves its cluster no rules after it restarts, until its token is fixed
