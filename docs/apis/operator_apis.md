<!-- Generator: Widdershins v4.0.1 -->

<h1 id="lobster-operator-apis-document">Lobster Operator APIs document v1.0</h1>

> Scroll down for example requests and responses.

Descriptions of Lobster log-sink management APIs

<h1 id="lobster-operator-apis-document-get">Get</h1>

## List sinks

`GET /api/v1/namespaces/{namespace}/sinks/{name}/{type}`

<h3 id="list-sinks-parameters">Parameters</h3>

|Name|In|Type|Required|Description|
|---|---|---|---|---|
|namespace|path|string|true|namespace name|
|name|path|string|true|sink name|
|type|path|string|true|deprecated;sink type (logMetricRules, logExportRules)|
|type|query|string|false|sink type (logMetricRules, logExportRules)|

> Example responses

> 200 Response

```json
[
  {
    "description": "string",
    "logExportRules": [
      {
        "basicBucket": {
          "destination": "string",
          "rootPath": "string",
          "timeLayoutOfSubDirectory": "2006-01"
        },
        "description": "string",
        "filter": {
          "clusters": [
            "string"
          ],
          "containers": [
            "string"
          ],
          "exclude": "string",
          "include": "string",
          "labels": [
            {
              "property1": "string",
              "property2": "string"
            }
          ],
          "namespace": "string",
          "pods": [
            "string"
          ],
          "setNames": [
            "string"
          ],
          "sources": [
            {
              "path": "string",
              "type": "string"
            }
          ]
        },
        "interval": "time duration(e.g. 1m)",
        "name": "string",
        "s3Bucket": {
          "accessKey": "string",
          "bucketName": "string",
          "destination": "string",
          "region": "string",
          "rootPath": "string",
          "secretKey": "string",
          "tags": {
            "property1": "string",
            "property2": "string"
          },
          "timeLayoutOfSubDirectory": "2006-01"
        }
      }
    ],
    "logMetricRules": [
      {
        "description": "string",
        "filter": {
          "clusters": [
            "string"
          ],
          "containers": [
            "string"
          ],
          "exclude": "string",
          "include": "string",
          "labels": [
            {
              "property1": "string",
              "property2": "string"
            }
          ],
          "namespace": "string",
          "pods": [
            "string"
          ],
          "setNames": [
            "string"
          ],
          "sources": [
            {
              "path": "string",
              "type": "string"
            }
          ]
        },
        "name": "string"
      }
    ],
    "name": "string",
    "namespace": "string",
    "type": "string"
  }
]
```

<h3 id="list-sinks-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|OK|Inline|
|400|[Bad Request](https://tools.ietf.org/html/rfc7231#section-6.5.1)|Invalid parameters|string|
|405|[Method Not Allowed](https://tools.ietf.org/html/rfc7231#section-6.5.5)|Method not allowed|string|
|500|[Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1)|Failed to get sink|string|

<h3 id="list-sinks-responseschema">Response Schema</h3>

Status Code **200**

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|*anonymous*|[[v1.Sink](#schemav1.sink)]|false|none|none|
|»» interval|string|false|none|Interval to export logs|
|»» name|string|false|none|Rule name|
|» logMetricRules|[[v1.LogMetricRule](#schemav1.logmetricrule)]|false|none|none|
|» name|string|false|none|none|
|» namespace|string|false|none|none|
|» type|string|false|none|none|

<aside class="success">
This operation does not require authentication
</aside>

<h1 id="lobster-operator-apis-document-put">Put</h1>

## Put log sink

`PUT /api/v1/namespaces/{namespace}/sinks/{name}/{type}`

> Body parameter

```json
{
  "description": "string",
  "logExportRules": [
    {
      "basicBucket": {
        "destination": "string",
        "rootPath": "string",
        "timeLayoutOfSubDirectory": "2006-01"
      },
      "description": "string",
      "filter": {
        "clusters": [
          "string"
        ],
        "containers": [
          "string"
        ],
        "exclude": "string",
        "include": "string",
        "labels": [
          {
            "property1": "string",
            "property2": "string"
          }
        ],
        "namespace": "string",
        "pods": [
          "string"
        ],
        "setNames": [
          "string"
        ],
        "sources": [
          {
            "path": "string",
            "type": "string"
          }
        ]
      },
      "interval": "time duration(e.g. 1m)",
      "name": "string",
      "s3Bucket": {
        "accessKey": "string",
        "bucketName": "string",
        "destination": "string",
        "region": "string",
        "rootPath": "string",
        "secretKey": "string",
        "tags": {
          "property1": "string",
          "property2": "string"
        },
        "timeLayoutOfSubDirectory": "2006-01"
      }
    }
  ],
  "logMetricRules": [
    {
      "description": "string",
      "filter": {
        "clusters": [
          "string"
        ],
        "containers": [
          "string"
        ],
        "exclude": "string",
        "include": "string",
        "labels": [
          {
            "property1": "string",
            "property2": "string"
          }
        ],
        "namespace": "string",
        "pods": [
          "string"
        ],
        "setNames": [
          "string"
        ],
        "sources": [
          {
            "path": "string",
            "type": "string"
          }
        ]
      },
      "name": "string"
    }
  ],
  "name": "string",
  "namespace": "string",
  "type": "string"
}
```

<h3 id="put-log-sink-parameters">Parameters</h3>

|Name|In|Type|Required|Description|
|---|---|---|---|---|
|namespace|path|string|true|namespace name|
|name|path|string|true|sink name|
|type|path|string|true|deprecated;sink type (logMetricRules, logExportRules)|
|body|body|[v1.Sink](#schemav1.sink)|true|sink contentd; Each content in array must be unique|

> Example responses

> 200 Response

<h3 id="put-log-sink-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|OK|string|
|201|[Created](https://tools.ietf.org/html/rfc7231#section-6.3.2)|Created successfully|string|
|400|[Bad Request](https://tools.ietf.org/html/rfc7231#section-6.5.1)|Invalid parameters|string|
|405|[Method Not Allowed](https://tools.ietf.org/html/rfc7231#section-6.5.5)|Method not allowed|string|
|422|[Unprocessable Entity](https://tools.ietf.org/html/rfc2518#section-10.3)|Restricted by limits|string|
|500|[Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1)|Failed to get sink content|string|

<aside class="success">
This operation does not require authentication
</aside>

<h1 id="lobster-operator-apis-document-delete">Delete</h1>

## Delete sink

`DELETE /api/v1/namespaces/{namespace}/sinks/{name}/rules/{rule}`

<h3 id="delete-sink-parameters">Parameters</h3>

|Name|In|Type|Required|Description|
|---|---|---|---|---|
|namespace|path|string|true|namespace name|
|name|path|string|true|sink name|
|rule|path|string|true|log export rule name to delete|
|ruleName|query|string|false|deprecated;metric rule name to delete|
|bucketName|query|string|false|deprecated;bucket name to delete|

> Example responses

> 200 Response

<h3 id="delete-sink-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|OK|string|
|400|[Bad Request](https://tools.ietf.org/html/rfc7231#section-6.5.1)|Invalid parameters|string|
|404|[Not Found](https://tools.ietf.org/html/rfc7231#section-6.5.4)|Not found|string|
|405|[Method Not Allowed](https://tools.ietf.org/html/rfc7231#section-6.5.5)|Method not allowed|string|
|500|[Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1)|Failed to delete sink|string|

<aside class="success">
This operation does not require authentication
</aside>

# Schemas

<h2 id="tocS_v1.BasicBucket">v1.BasicBucket</h2>
<!-- backwards compatibility -->
<a id="schemav1.basicbucket"></a>
<a id="schema_v1.BasicBucket"></a>
<a id="tocSv1.basicbucket"></a>
<a id="tocsv1.basicbucket"></a>

```json
{
  "destination": "string",
  "rootPath": "string",
  "timeLayoutOfSubDirectory": "2006-01"
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|destination|string|false|none|Address to export logs|
|rootPath|string|false|none|Root directory to store logs within external storage|
|timeLayoutOfSubDirectory|string|false|none|An option(default `2006-01`) that sets the name of the sub-directory following `{Root path}` to a time-based layout|

<h2 id="tocS_v1.Filter">v1.Filter</h2>
<!-- backwards compatibility -->
<a id="schemav1.filter"></a>
<a id="schema_v1.Filter"></a>
<a id="tocSv1.filter"></a>
<a id="tocsv1.filter"></a>

```json
{
  "clusters": [
    "string"
  ],
  "containers": [
    "string"
  ],
  "exclude": "string",
  "include": "string",
  "labels": [
    {
      "property1": "string",
      "property2": "string"
    }
  ],
  "namespace": "string",
  "pods": [
    "string"
  ],
  "setNames": [
    "string"
  ],
  "sources": [
    {
      "path": "string",
      "type": "string"
    }
  ]
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|clusters|[string]|false|none|Filter logs only for specific Clusters|
|containers|[string]|false|none|Filter logs only for specific Containers|
|exclude|string|false|none|Filter only logs that do not match the re2 expression(https://github.com/google/re2/wiki/Syntax)|
|include|string|false|none|Filter only logs that match the re2 expression(https://github.com/google/re2/wiki/Syntax)|
|labels|[object]|false|none|Filter logs only for specific Pod labels|
|» **additionalProperties**|string|false|none|none|
|namespace|string|false|none|Filter logs only for specific Namespace|
|pods|[string]|false|none|Filter logs only for specific Pods|
|setNames|[string]|false|none|Filter logs only for specific ReplicaSets/StatefulSets|
|sources|[[v1.Source](#schemav1.source)]|false|none|Filter logs only for specific Sources|

<h2 id="tocS_v1.LogExportRule">v1.LogExportRule</h2>
<!-- backwards compatibility -->
<a id="schemav1.logexportrule"></a>
<a id="schema_v1.LogExportRule"></a>
<a id="tocSv1.logexportrule"></a>
<a id="tocsv1.logexportrule"></a>

```json
{
  "basicBucket": {
    "destination": "string",
    "rootPath": "string",
    "timeLayoutOfSubDirectory": "2006-01"
  },
  "description": "string",
  "filter": {
    "clusters": [
      "string"
    ],
    "containers": [
      "string"
    ],
    "exclude": "string",
    "include": "string",
    "labels": [
      {
        "property1": "string",
        "property2": "string"
      }
    ],
    "namespace": "string",
    "pods": [
      "string"
    ],
    "setNames": [
      "string"
    ],
    "sources": [
      {
        "path": "string",
        "type": "string"
      }
    ]
  },
  "interval": "time duration(e.g. 1m)",
  "name": "string",
  "s3Bucket": {
    "accessKey": "string",
    "bucketName": "string",
    "destination": "string",
    "region": "string",
    "rootPath": "string",
    "secretKey": "string",
    "tags": {
      "property1": "string",
      "property2": "string"
    },
    "timeLayoutOfSubDirectory": "2006-01"
  }
}

```

### Properties

*None*

<h2 id="tocS_v1.LogMetricRule">v1.LogMetricRule</h2>
<!-- backwards compatibility -->
<a id="schemav1.logmetricrule"></a>
<a id="schema_v1.LogMetricRule"></a>
<a id="tocSv1.logmetricrule"></a>
<a id="tocsv1.logmetricrule"></a>

```json
{
  "description": "string",
  "filter": {
    "clusters": [
      "string"
    ],
    "containers": [
      "string"
    ],
    "exclude": "string",
    "include": "string",
    "labels": [
      {
        "property1": "string",
        "property2": "string"
      }
    ],
    "namespace": "string",
    "pods": [
      "string"
    ],
    "setNames": [
      "string"
    ],
    "sources": [
      {
        "path": "string",
        "type": "string"
      }
    ]
  },
  "name": "string"
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|description|string|false|none|Description of this rule|

<h2 id="tocS_v1.S3Bucket">v1.S3Bucket</h2>
<!-- backwards compatibility -->
<a id="schemav1.s3bucket"></a>
<a id="schema_v1.S3Bucket"></a>
<a id="tocSv1.s3bucket"></a>
<a id="tocsv1.s3bucket"></a>

```json
{
  "accessKey": "string",
  "bucketName": "string",
  "destination": "string",
  "region": "string",
  "rootPath": "string",
  "secretKey": "string",
  "tags": {
    "property1": "string",
    "property2": "string"
  },
  "timeLayoutOfSubDirectory": "2006-01"
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|accessKey|string|false|none|S3 bucket access key|
|bucketName|string|false|none|S3 bucket name|
|destination|string|false|none|S3 Address to export logs|
|region|string|false|none|S3 region|
|rootPath|string|false|none|Root directory to store logs within external storage|
|secretKey|string|false|none|S3 bucket secret key|

<h2 id="tocS_v1.Sink">v1.Sink</h2>
<!-- backwards compatibility -->
<a id="schemav1.sink"></a>
<a id="schema_v1.Sink"></a>
<a id="tocSv1.sink"></a>
<a id="tocsv1.sink"></a>

```json
{
  "description": "string",
  "logExportRules": [
    {
      "basicBucket": {
        "destination": "string",
        "rootPath": "string",
        "timeLayoutOfSubDirectory": "2006-01"
      },
      "description": "string",
      "filter": {
        "clusters": [
          "string"
        ],
        "containers": [
          "string"
        ],
        "exclude": "string",
        "include": "string",
        "labels": [
          {
            "property1": "string",
            "property2": "string"
          }
        ],
        "namespace": "string",
        "pods": [
          "string"
        ],
        "setNames": [
          "string"
        ],
        "sources": [
          {
            "path": "string",
            "type": "string"
          }
        ]
      },
      "interval": "time duration(e.g. 1m)",
      "name": "string",
      "s3Bucket": {
        "accessKey": "string",
        "bucketName": "string",
        "destination": "string",
        "region": "string",
        "rootPath": "string",
        "secretKey": "string",
        "tags": {
          "property1": "string",
          "property2": "string"
        },
        "timeLayoutOfSubDirectory": "2006-01"
      }
    }
  ],
  "logMetricRules": [
    {
      "description": "string",
      "filter": {
        "clusters": [
          "string"
        ],
        "containers": [
          "string"
        ],
        "exclude": "string",
        "include": "string",
        "labels": [
          {
            "property1": "string",
            "property2": "string"
          }
        ],
        "namespace": "string",
        "pods": [
          "string"
        ],
        "setNames": [
          "string"
        ],
        "sources": [
          {
            "path": "string",
            "type": "string"
          }
        ]
      },
      "name": "string"
    }
  ],
  "name": "string",
  "namespace": "string",
  "type": "string"
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|description|string|false|none|none|
|logExportRules|[[v1.LogExportRule](#schemav1.logexportrule)]|false|none|none|

<h2 id="tocS_v1.Source">v1.Source</h2>
<!-- backwards compatibility -->
<a id="schemav1.source"></a>
<a id="schema_v1.Source"></a>
<a id="tocSv1.source"></a>
<a id="tocsv1.source"></a>

```json
{
  "path": "string",
  "type": "string"
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|path|string|false|none|none|
|type|string|false|none|none|

<h2 id="tocS_v1.Tags">v1.Tags</h2>
<!-- backwards compatibility -->
<a id="schemav1.tags"></a>
<a id="schema_v1.Tags"></a>
<a id="tocSv1.tags"></a>
<a id="tocsv1.tags"></a>

```json
{
  "property1": "string",
  "property2": "string"
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|**additionalProperties**|string|false|none|none|

