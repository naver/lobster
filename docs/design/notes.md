## Issues associated with asynchronous logging
- Multiple applications, including Envoy, can perform logging operations asynchronously.
- The following issues have been identified in association with asynchronous logging, along with the measures taken to address them.

### Out of timestamp order

- Issue description
  - There may be a discrepancy between the timestamp in the log and the time of logging, resulting in inconsistencies such as querying logs outside their intended time period
- Measures taken
  - Logs are collected and sorted in chronological order within a specified time period before being stored
    - This operation is performed only within the `store.leakyBucketInterval (default 1s)`
    - Logs with a timestamp difference greater than this interval are written directly without reordering
  - This issue can occur if logs arrive in reverse order during the collection and storage cycle

### Explanation of each file log rotation method

- Issue description
  - `rename & create`: Before the logger's pointer moves to a new log file, a portion of the latest log may be recorded in the rotated file
  - `copy & truncate`: If truncation happens before the log is read, log data loss occurs
  - Reference: https://grafana.com/docs/loki/latest/send-data/promtail/logrotation/
- Measures taken
  - For `rename & create`, the latest log is kept in the rotated file, so a feature to track it in the rotated file for a certain amount of time has been implemented
    - If logging continues to the rotated file after a certain amount of time, those logs may be ignored
  - For `copy & truncate`, this issue involves the disappearance of traceable logs, so log tracking is not supported
    - It is recommended to use `rename & create` for log rotation

