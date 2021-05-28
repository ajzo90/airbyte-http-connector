
#### Build images

```
docker build -t airbyte/source-go-example:dev -f Dockerfile-source .
docker build -t airbyte/destination-go-example:dev -f Dockerfile-destination .
docker build -t airbyte/server:dev -f Dockerfile-server .
```

#### Start test server
The dummy test server support both source and destination protocols. 
The source is hard-coded to generate 100000 dummy records for 3 streams with different properties. 
The destination parses the input stream, and just count the number of records. 
```
docker run --rm -p 9092:9092 airbyte/server:dev
```

#### Setup airbyte (source def, target def, source, target)
Use UI,API or docker.
```
docker build -t airbyte/setup:dev -f Dockerfile-setup .
docker run --rm --network host airbyte/setup:dev # "works on my maching..." 
```

```
Created destination definition {dest-def airbyte/destination-go-example dev example.com  d2961419-feec-4806-8b90-ded07551b293}
Created source definition {src-def airbyte/source-go-example dev example.com cd5d4b78-b393-4d05-89a9-d8f4946f7792 }
Created destination {dest  d2961419-feec-4806-8b90-ded07551b293 {http://127.0.0.1:9092 secret} 5ae6b09b-fdec-41af-aaf7-7d94cfc33ef6 }
Created source {src cd5d4b78-b393-4d05-89a9-d8f4946f7792  {http://127.0.0.1:9092 secret} 5ae6b09b-fdec-41af-aaf7-7d94cfc33ef6 61a64c01-611e-4e50-89db-ec09c62f6b0e}
```

#### Use source or/and target in your airbyte setup


### example requests
#### check
```
% curl -s -d '{"key":"secret"}' http://127.0.0.1:9092/source/check
{"type":"CONNECTION_STATUS","connectionStatus":{"status":"SUCCEEDED"}}
```

#### check (failed)
```
% curl -s -d '{"key":"secr"}' http://127.0.0.1:9092/source/check 
{"type":"CONNECTION_STATUS","connectionStatus":{"status":"FAILED","message":"invalid Key"}}
```

#### discover (first stream output)
```
% curl -s -d '{"key":"secret"}' http://127.0.0.1:9092/source/discover | jq '.catalog.streams | .[0]'
{
  "name": "table1",
  "namespace": "app1",
  "json_schema": {
    "$schema": "http://json-schema.org/draft-07/schema#",
    "type": "object",
    "properties": {
      "id": {
        "description": "hello",
        "type": "int"
      },
      "value": {
        "description": "world",
        "type": "int"
      }
    }
  },
  "supported_sync_modes": [
    "full_refresh",
    "incremental"
  ],
  "source_defined_primary_key": [
    [
      "id"
    ]
  ]
}
```

#### read
```
source-go-example % curl -s -d @readConfig.json http://127.0.0.1:9092/source/read | head
{"type":"RECORD","record":{"stream":"table1","data":{"id":0,"v":0},"emitted_at":1622208594225}}
{"type":"RECORD","record":{"stream":"table1","data":{"id":1,"v":123},"emitted_at":1622208594225}}
{"type":"RECORD","record":{"stream":"table1","data":{"id":2,"v":246},"emitted_at":1622208594225}}
{"type":"RECORD","record":{"stream":"table1","data":{"id":3,"v":369},"emitted_at":1622208594225}}
{"type":"RECORD","record":{"stream":"table1","data":{"id":4,"v":492},"emitted_at":1622208594225}}
{"type":"RECORD","record":{"stream":"table1","data":{"id":5,"v":615},"emitted_at":1622208594225}}
{"type":"RECORD","record":{"stream":"table1","data":{"id":6,"v":738},"emitted_at":1622208594225}}
{"type":"RECORD","record":{"stream":"table1","data":{"id":7,"v":861},"emitted_at":1622208594225}}
{"type":"RECORD","record":{"stream":"table1","data":{"id":8,"v":984},"emitted_at":1622208594225}}
{"type":"RECORD","record":{"stream":"table1","data":{"id":9,"v":1107},"emitted_at":1622208594225}}
```

#### write
```
curl -F 'config={"config": {"key": "secret"}}' -F 'data=@stream.txt' http://127.0.0.1:9092/destination/write
```