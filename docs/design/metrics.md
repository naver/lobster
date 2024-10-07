## Supported metrics

To visualize log data, it is provided in [Prometheus metric format](https://prometheus.io/docs/concepts/data_model/).

### Metric names

- `Log collection`: This metric represents the amount logs being collected and occurrences of logs exceeding the limit
- `Log sink`: This metric represents the logs associated with Log Metric/Export
- `Loggen`: This metric represents the test results from [Loggen](./loggen.md)

Category | Name | Type | Description
--- | --- | --- | ---
`Log collection` | `lobster_tailed_lines_total` | `Counter` | Number of log lines collected
`Log collection` | `lobster_tailed_bytes_total` | `Counter` | Log size collected
`Log collection` | `lobster_overloaded_target_total` | `Counter` | Occurs when logs are restricted due to high volumes
`Log sink` | `lobster_log_sink_bytes_total` | `Counter` | Log size measured per unit of log sink (export) 
`Log sink` | `lobster_log_sink_failure_total` | `Counter` | Log sink failure (e.g., destination timeout, invalid regexp)
`Log sink` | `lobster_log_metric_matched_logs_total` | `Counter` | Log occurrences accumulated based on the log sink (metric) rules
`Loggen` | `lobster_loggen_failure_total` | `Counter` | A count of failure of inspection
`Loggen` | `lobster_loggen_verified` | `Gauge` | A count of verified logs
`Loggen` | `lobster_loggen_appeared_time_seconds` | `Gauge` | The time it takes to reflect the latest logs

### Metric labels

Name | Description
--- | ---
`target_namespace` | Namespace to collect metrics
`sink_name` | Log sink (custom resource) name
`sink_namespace` | Namespace where log sink (custom resource) is installed
`sink_type` | Log sink types (e.g., logMetricRules...)
`sink_contents_name` | Configuration name such as export, metrics, which is defined in the log sink
`log_namespace` | Namespace to which the container generating logs belongs
`log_pod` | Pod to which the container generating logs belongs
`log_container` | Container that generates logs
`log_source_type` | Types of logs generated in the container (stdstream, emptydir)
`log_source_path` | Log path information for log types in emptyDir (`/` is replaced by `_`.)