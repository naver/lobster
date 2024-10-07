<!-- Generator: Widdershins v4.0.1 -->

<h1 id="lobster-api-document">Lobster API document v1.0</h1>

> Scroll down for example requests and responses.

Descriptions of Lobster global query APIs

<h1 id="lobster-api-document-post">Post</h1>

## Get metadata of logs

`POST /api/{version}/logs`

Get metadata of logs for conditions

> Body parameter

```json
{
  "attachment": false,
  "burst": 0,
  "clusters": [
    "string"
  ],
  "container": "string",
  "containers": [
    "string"
  ],
  "end": "string",
  "exclude": "string",
  "id": "string",
  "include": "string",
  "labels": [
    {
      "property1": "string",
      "property2": "string"
    }
  ],
  "local": false,
  "namespace": "string",
  "namespaces": [
    "string"
  ],
  "page": 0,
  "pod": "string",
  "pod_uid": "string",
  "pods": [
    "string"
  ],
  "setName": "string",
  "setNames": [
    "string"
  ],
  "source": {
    "path": "string",
    "type": "string"
  },
  "sources": [
    {
      "path": "string",
      "type": "string"
    }
  ],
  "start": "string"
}
```

<h3 id="get-metadata-of-logs-parameters">Parameters</h3>

|Name|In|Type|Required|Description|
|---|---|---|---|---|
|version|path|string|true|v1 or v2|
|body|body|[query.Request](#schemaquery.request)|true|request parameters|

> Example responses

> 200 Response

```json
[
  {
    "cluster": "string",
    "container": "string",
    "id": "string",
    "labels": {
      "property1": "string",
      "property2": "string"
    },
    "line": 0,
    "namespace": "string",
    "pod": "string",
    "podUid": "string",
    "setName": "string",
    "size": 0,
    "source": {
      "path": "string",
      "type": "string"
    },
    "startedAt": "string",
    "storeAddr": "string",
    "updatedAt": "string"
  }
]
```

<h3 id="get-metadata-of-logs-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|OK|Inline|
|204|[No Content](https://tools.ietf.org/html/rfc7231#section-6.3.5)|No chunks|string|
|400|[Bad Request](https://tools.ietf.org/html/rfc7231#section-6.5.1)|Invalid parameters|string|
|405|[Method Not Allowed](https://tools.ietf.org/html/rfc7231#section-6.5.5)|Method not allowed|string|
|429|[Too Many Requests](https://tools.ietf.org/html/rfc6585#section-4)|too many requests|string|
|500|[Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1)|Failed to read logs|string|

<h3 id="get-metadata-of-logs-responseschema">Response Schema</h3>

Status Code **200**

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|*anonymous*|[[model.Chunk](#schemamodel.chunk)]|false|none|none|

<aside class="success">
This operation does not require authentication
</aside>

## Export logs within range

`POST /api/{version}/logs/export`

Export logs for conditions

> Body parameter

```json
{
  "attachment": false,
  "burst": 0,
  "clusters": [
    "string"
  ],
  "container": "string",
  "containers": [
    "string"
  ],
  "end": "string",
  "exclude": "string",
  "id": "string",
  "include": "string",
  "labels": [
    {
      "property1": "string",
      "property2": "string"
    }
  ],
  "local": false,
  "namespace": "string",
  "namespaces": [
    "string"
  ],
  "page": 0,
  "pod": "string",
  "pod_uid": "string",
  "pods": [
    "string"
  ],
  "setName": "string",
  "setNames": [
    "string"
  ],
  "source": {
    "path": "string",
    "type": "string"
  },
  "sources": [
    {
      "path": "string",
      "type": "string"
    }
  ],
  "start": "string"
}
```

<h3 id="export-logs-within-range-parameters">Parameters</h3>

|Name|In|Type|Required|Description|
|---|---|---|---|---|
|version|path|string|true|v1 or v2|
|body|body|[query.Request](#schemaquery.request)|true|request parameters|

> Example responses

> 200 Response

```json
"string"
```

<h3 id="export-logs-within-range-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|OK|string|
|204|[No Content](https://tools.ietf.org/html/rfc7231#section-6.3.5)|No chunks|string|
|400|[Bad Request](https://tools.ietf.org/html/rfc7231#section-6.5.1)|Invalid parameters|string|
|405|[Method Not Allowed](https://tools.ietf.org/html/rfc7231#section-6.5.5)|Method not allowed|string|
|429|[Too Many Requests](https://tools.ietf.org/html/rfc6585#section-4)|too many requests|string|
|500|[Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1)|Failed to read logs|string|

<aside class="success">
This operation does not require authentication
</aside>

## Get series within range

`POST /api/{version}/logs/series`

Get series for conditions

> Body parameter

```json
{
  "attachment": false,
  "burst": 0,
  "clusters": [
    "string"
  ],
  "container": "string",
  "containers": [
    "string"
  ],
  "end": "string",
  "exclude": "string",
  "id": "string",
  "include": "string",
  "labels": [
    {
      "property1": "string",
      "property2": "string"
    }
  ],
  "local": false,
  "namespace": "string",
  "namespaces": [
    "string"
  ],
  "page": 0,
  "pod": "string",
  "pod_uid": "string",
  "pods": [
    "string"
  ],
  "setName": "string",
  "setNames": [
    "string"
  ],
  "source": {
    "path": "string",
    "type": "string"
  },
  "sources": [
    {
      "path": "string",
      "type": "string"
    }
  ],
  "start": "string"
}
```

<h3 id="get-series-within-range-parameters">Parameters</h3>

|Name|In|Type|Required|Description|
|---|---|---|---|---|
|version|path|string|true|v1 or v2|
|body|body|[query.Request](#schemaquery.request)|true|request parameters|

> Example responses

> 200 Response

```json
{
  "contents": "string",
  "pageInfo": {
    "current": 0,
    "hasNext": true,
    "isPartialContents": true,
    "total": 0
  },
  "series": [
    {
      "chunk_key": "string",
      "lines": 0,
      "name": "string",
      "samples": [
        {
          "lines": 0,
          "size": 0,
          "timestamp": "string"
        }
      ],
      "size": 0
    }
  ]
}
```

<h3 id="get-series-within-range-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|OK|[query.Response](#schemaquery.response)|
|204|[No Content](https://tools.ietf.org/html/rfc7231#section-6.3.5)|No chunks|string|
|400|[Bad Request](https://tools.ietf.org/html/rfc7231#section-6.5.1)|Invalid parameters|string|
|405|[Method Not Allowed](https://tools.ietf.org/html/rfc7231#section-6.5.5)|Method not allowed|string|
|429|[Too Many Requests](https://tools.ietf.org/html/rfc6585#section-4)|too many requests|string|
|500|[Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1)|Failed to read logs|string|

<aside class="success">
This operation does not require authentication
</aside>

## Get logs within range

`POST /api/v2/logs/range`

Get logs for conditions

> Body parameter

```json
{
  "attachment": false,
  "burst": 0,
  "clusters": [
    "string"
  ],
  "container": "string",
  "containers": [
    "string"
  ],
  "end": "string",
  "exclude": "string",
  "id": "string",
  "include": "string",
  "labels": [
    {
      "property1": "string",
      "property2": "string"
    }
  ],
  "local": false,
  "namespace": "string",
  "namespaces": [
    "string"
  ],
  "page": 0,
  "pod": "string",
  "pod_uid": "string",
  "pods": [
    "string"
  ],
  "setName": "string",
  "setNames": [
    "string"
  ],
  "source": {
    "path": "string",
    "type": "string"
  },
  "sources": [
    {
      "path": "string",
      "type": "string"
    }
  ],
  "start": "string"
}
```

<h3 id="get-logs-within-range-parameters">Parameters</h3>

|Name|In|Type|Required|Description|
|---|---|---|---|---|
|body|body|[query.Request](#schemaquery.request)|true|request parameters|

> Example responses

> 200 Response

```json
{
  "contents": [
    {
      "cluster": "string",
      "container": "string",
      "labels": {
        "property1": "string",
        "property2": "string"
      },
      "message": "string",
      "namespace": "string",
      "pod": "string",
      "sourcePath": "string",
      "sourceType": "string",
      "stream": "string",
      "tag": "string",
      "time": "string"
    }
  ],
  "pageInfo": {
    "current": 0,
    "hasNext": true,
    "isPartialContents": true,
    "total": 0
  },
  "series": [
    {
      "chunk_key": "string",
      "lines": 0,
      "name": "string",
      "samples": [
        {
          "lines": 0,
          "size": 0,
          "timestamp": "string"
        }
      ],
      "size": 0
    }
  ]
}
```

<h3 id="get-logs-within-range-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|OK|[query.ResponseEntries](#schemaquery.responseentries)|
|204|[No Content](https://tools.ietf.org/html/rfc7231#section-6.3.5)|No chunks|string|
|400|[Bad Request](https://tools.ietf.org/html/rfc7231#section-6.5.1)|Invalid parameters|string|
|405|[Method Not Allowed](https://tools.ietf.org/html/rfc7231#section-6.5.5)|Method not allowed|string|
|429|[Too Many Requests](https://tools.ietf.org/html/rfc6585#section-4)|too many requests|string|
|500|[Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1)|Failed to read logs|string|
|501|[Not Implemented](https://tools.ietf.org/html/rfc7231#section-6.6.2)|Not supported version|string|

<aside class="success">
This operation does not require authentication
</aside>

# Schemas

<h2 id="tocS_github_com_naver_lobster_pkg_lobster_model.Sample">github_com_naver_lobster_pkg_lobster_model.Sample</h2>
<!-- backwards compatibility -->
<a id="schemagithub_com_naver_lobster_pkg_lobster_model.sample"></a>
<a id="schema_github_com_naver_lobster_pkg_lobster_model.Sample"></a>
<a id="tocSgithub_com_naver_lobster_pkg_lobster_model.sample"></a>
<a id="tocsgithub_com_naver_lobster_pkg_lobster_model.sample"></a>

```json
{
  "lines": 0,
  "size": 0,
  "timestamp": "string"
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|lines|integer|false|none|none|
|size|integer|false|none|none|
|timestamp|string|false|none|none|

<h2 id="tocS_model.Chunk">model.Chunk</h2>
<!-- backwards compatibility -->
<a id="schemamodel.chunk"></a>
<a id="schema_model.Chunk"></a>
<a id="tocSmodel.chunk"></a>
<a id="tocsmodel.chunk"></a>

```json
{
  "cluster": "string",
  "container": "string",
  "id": "string",
  "labels": {
    "property1": "string",
    "property2": "string"
  },
  "line": 0,
  "namespace": "string",
  "pod": "string",
  "podUid": "string",
  "setName": "string",
  "size": 0,
  "source": {
    "path": "string",
    "type": "string"
  },
  "startedAt": "string",
  "storeAddr": "string",
  "updatedAt": "string"
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|cluster|string|false|none|none|
|container|string|false|none|none|
|id|string|false|none|none|

<h2 id="tocS_model.Entry">model.Entry</h2>
<!-- backwards compatibility -->
<a id="schemamodel.entry"></a>
<a id="schema_model.Entry"></a>
<a id="tocSmodel.entry"></a>
<a id="tocsmodel.entry"></a>

```json
{
  "cluster": "string",
  "container": "string",
  "labels": {
    "property1": "string",
    "property2": "string"
  },
  "message": "string",
  "namespace": "string",
  "pod": "string",
  "sourcePath": "string",
  "sourceType": "string",
  "stream": "string",
  "tag": "string",
  "time": "string"
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|cluster|string|false|none|none|
|container|string|false|none|none|
|labels|object|false|none|none|
|» **additionalProperties**|string|false|none|none|
|message|string|false|none|none|
|namespace|string|false|none|none|
|pod|string|false|none|none|
|sourcePath|string|false|none|none|
|sourceType|string|false|none|none|
|stream|string|false|none|none|
|tag|string|false|none|none|
|time|string|false|none|none|

<h2 id="tocS_model.Labels">model.Labels</h2>
<!-- backwards compatibility -->
<a id="schemamodel.labels"></a>
<a id="schema_model.Labels"></a>
<a id="tocSmodel.labels"></a>
<a id="tocsmodel.labels"></a>

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

<h2 id="tocS_model.PageInfo">model.PageInfo</h2>
<!-- backwards compatibility -->
<a id="schemamodel.pageinfo"></a>
<a id="schema_model.PageInfo"></a>
<a id="tocSmodel.pageinfo"></a>
<a id="tocsmodel.pageinfo"></a>

```json
{
  "current": 0,
  "hasNext": true,
  "isPartialContents": true,
  "total": 0
}

```

Page inforamtion.

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|current|integer|false|none|none|
|hasNext|boolean|false|none|none|
|isPartialContents|boolean|false|none|partial logs are returned|
|total|integer|false|none|none|

<h2 id="tocS_model.Series">model.Series</h2>
<!-- backwards compatibility -->
<a id="schemamodel.series"></a>
<a id="schema_model.Series"></a>
<a id="tocSmodel.series"></a>
<a id="tocsmodel.series"></a>

```json
{
  "chunk_key": "string",
  "lines": 0,
  "name": "string",
  "samples": [
    {
      "lines": 0,
      "size": 0,
      "timestamp": "string"
    }
  ],
  "size": 0
}

```

Name: "{cluster}_{namespace}_{pod}_{container}_{source}-{file number}".

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|chunk_key|string|false|none|none|
|lines|integer|false|none|none|
|name|string|false|none|none|
|samples|[[github_com_naver_lobster_pkg_lobster_model.Sample](#schemagithub_com_naver_lobster_pkg_lobster_model.sample)]|false|none|none|

<h2 id="tocS_model.Source">model.Source</h2>
<!-- backwards compatibility -->
<a id="schemamodel.source"></a>
<a id="schema_model.Source"></a>
<a id="tocSmodel.source"></a>
<a id="tocsmodel.source"></a>

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

<h2 id="tocS_query.Request">query.Request</h2>
<!-- backwards compatibility -->
<a id="schemaquery.request"></a>
<a id="schema_query.Request"></a>
<a id="tocSquery.request"></a>
<a id="tocsquery.request"></a>

```json
{
  "attachment": false,
  "burst": 0,
  "clusters": [
    "string"
  ],
  "container": "string",
  "containers": [
    "string"
  ],
  "end": "string",
  "exclude": "string",
  "id": "string",
  "include": "string",
  "labels": [
    {
      "property1": "string",
      "property2": "string"
    }
  ],
  "local": false,
  "namespace": "string",
  "namespaces": [
    "string"
  ],
  "page": 0,
  "pod": "string",
  "pod_uid": "string",
  "pods": [
    "string"
  ],
  "setName": "string",
  "setNames": [
    "string"
  ],
  "source": {
    "path": "string",
    "type": "string"
  },
  "sources": [
    {
      "path": "string",
      "type": "string"
    }
  ],
  "start": "string"
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|attachment|boolean|false|none|none|
|burst|integer|false|none|The number of logs that can be returned in one page and this can be greater or less than burst|
|clusters|[string]|false|none|Get chunks belongs to clusters|
|container|string|false|none|Use internally|
|containers|[string]|false|none|Get chunks belongs to namespace and containers|
|end|string|false|none|End time for query|
|exclude|string|false|none|none|
|id|string|false|none|Use internally|
|include|string|false|none|Regular expression to search logs|
|labels|[[model.Labels](#schemamodel.labels)]|false|none|Get chunks belongs to namespaces and labels|

<h2 id="tocS_query.Response">query.Response</h2>
<!-- backwards compatibility -->
<a id="schemaquery.response"></a>
<a id="schema_query.Response"></a>
<a id="tocSquery.response"></a>
<a id="tocsquery.response"></a>

```json
{
  "contents": "string",
  "pageInfo": {
    "current": 0,
    "hasNext": true,
    "isPartialContents": true,
    "total": 0
  },
  "series": [
    {
      "chunk_key": "string",
      "lines": 0,
      "name": "string",
      "samples": [
        {
          "lines": 0,
          "size": 0,
          "timestamp": "string"
        }
      ],
      "size": 0
    }
  ]
}

```

Response wrapping series and logs from store.

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|contents|string|false|none|logs in string|

<h2 id="tocS_query.ResponseEntries">query.ResponseEntries</h2>
<!-- backwards compatibility -->
<a id="schemaquery.responseentries"></a>
<a id="schema_query.ResponseEntries"></a>
<a id="tocSquery.responseentries"></a>
<a id="tocsquery.responseentries"></a>

```json
{
  "contents": [
    {
      "cluster": "string",
      "container": "string",
      "labels": {
        "property1": "string",
        "property2": "string"
      },
      "message": "string",
      "namespace": "string",
      "pod": "string",
      "sourcePath": "string",
      "sourceType": "string",
      "stream": "string",
      "tag": "string",
      "time": "string"
    }
  ],
  "pageInfo": {
    "current": 0,
    "hasNext": true,
    "isPartialContents": true,
    "total": 0
  },
  "series": [
    {
      "chunk_key": "string",
      "lines": 0,
      "name": "string",
      "samples": [
        {
          "lines": 0,
          "size": 0,
          "timestamp": "string"
        }
      ],
      "size": 0
    }
  ]
}

```

Response wrapping series and logs from querier.

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|contents|[[model.Entry](#schemamodel.entry)]|false|none|log entries|

