package airbyte

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

type SourceConfig struct {
	Url string `json:"url"`
	Key string `json:"key"`
}

type ReadConfig struct {
	Config  SourceConfig             `json:"config"`
	Catalog ConfiguredAirbyteCatalog `json:"catalog"`
	State   interface{}              `json:"state"`
}

type SourceStateWrap struct {
	Type  string          `json:"type"`
	State SourceStateData `json:"state"`
}
type SourceStateData struct {
	Data SourceState `json:"data"`
}

type SourceState struct {
	Hello string
	Info  string
}

func validateSourceConfig(cnf SourceConfig) error {
	if cnf.Key != "secret" {
		return fmt.Errorf("invalid Key")
	}
	return nil
}

func SourceRead(w io.Writer, cnf ReadConfig) error {
	if err := validateSourceConfig(cnf.Config); err != nil {
		return err
	}

	type Data struct {
		Id int `json:"id"`
		V  int `json:"v"`
	}
	type Record struct {
		Stream    string `json:"stream"`
		Data      Data   `json:"data"`
		EmittedAt int64  `json:"emitted_at"`
	}

	type RecordWrap struct {
		Type   string `json:"type"`
		Record Record `json:"record"`
	}

	for _, stream := range cnf.Catalog.Streams {
		var streamName = stream.Stream.Name
		var e = json.NewEncoder(w)
		var record = RecordWrap{
			Type: "RECORD",
			Record: Record{
				EmittedAt: time.Now().UnixNano() / 1000000,
				Stream:    streamName,
			},
		}

		for i := 0; i < 1e5; i++ {
			record.Record.Data = Data{Id: i, V: i * 123}
			if err := e.Encode(record); err != nil {
				return err
			}
		}

		if err := json.NewEncoder(w).Encode(SourceStateWrap{Type: "STATE", State: SourceStateData{Data: SourceState{Info: "test123", Hello: "world"}}}); err != nil {
			return err
		}
	}

	return nil
}

func SourceDiscover(w io.Writer, cnf SourceConfig) error {
	if err := validateSourceConfig(cnf); err != nil {
		return err
	}
	return json.NewEncoder(w).Encode(AirbyteCatalogWrapper{Type: "CATALOG",
		Catalog: AirbyteCatalog{
			Streams: []AirbyteStream{
				{
					Name:                    "table1",
					Namespace:               "app1",
					SupportedSyncModes:      []string{"full_refresh", "incremental"},
					SourceDefinedPrimaryKey: [][]string{{"id"}},
					Schema: JSONSchema{
						Schema: "http://json-schema.org/draft-07/schema#",
						Type:   "object",
						Properties: map[string]Property{
							"id":    {Type: "int", Description: "hello"},
							"value": {Type: "int", Description: "world"},
						},
					},
				},
				{
					Name:      "table2",
					Namespace: "app1",
					//SourceDefinedCursor: true,
					Schema: JSONSchema{
						Schema: "http://json-schema.org/draft-07/schema#",
						Type:   "object",
						Properties: map[string]Property{
							"id":    {Type: "int"},
							"value": {Type: "int"},
						},
					},
				},
				{
					Name:               "table3",
					Namespace:          "app1",
					SupportedSyncModes: []string{"full_refresh"},
					DefaultCursorField: []string{"id"},
					Schema: JSONSchema{
						Schema: "http://json-schema.org/draft-07/schema#",
						Type:   "object",
						Properties: map[string]Property{
							"id":    {Type: "int"},
							"value": {Type: "int"},
						},
					},
				},
			}}})
}

func SourceCheck(w io.Writer, cnf SourceConfig) error {
	var ret = ConnectionStatusWrap{Type: "CONNECTION_STATUS", ConnectionStatus: ConnectionStatus{Status: "SUCCEEDED"}}
	if err := validateSourceConfig(cnf); err != nil {
		ret.ConnectionStatus.Status = "FAILED"
		ret.ConnectionStatus.Message = err.Error()
	}
	return json.NewEncoder(w).Encode(ret)
}

func SourceSpec(w io.Writer) error {
	return json.NewEncoder(w).Encode(ConnectorSpecificationWrap{Type: "SPEC", Spec: ConnectorSpecification{
		DocumentationUrl:    "https://docs.airbyte.io/integrations/destinations/mssql",
		SupportsIncremental: true,
		ConnectionSpecification: ConnectionSpecification{
			Schema:               "http://json-schema.org/draft-07/schema#",
			Title:                "Airbyte HTTP Source Spec",
			Type:                 "object",
			Required:             []string{"url", "key"},
			AdditionalProperties: false,
			Properties: map[string]Property{
				"url": {Description: "host", Type: "string"},
				"key": {Description: "port", Type: "string"},
			},
		},
	}})
}

/*
spec() -> ConnectorSpecification
check(Config) -> AirbyteConnectionStatus
discover(Config) -> AirbyteCatalog
read(Config, AirbyteCatalog, State) -> Stream<AirbyteMessage>
*/

/*
docker run --rm -i <source-image-name> spec
docker run --rm -i <source-image-name> check --config <config-file-path>
docker run --rm -i <source-image-name> discover --config <config-file-path>
docker run --rm -i <source-image-name> read --config <config-file-path> --catalog <catalog-file-path> [--state <state-file-path>] > message_stream.json
*/

/* Spec
docker run --rm -i -v /Users/christianpersson/repos/airbyte:/tmp airbyte/source-file spec
{"type": "SPEC", "spec": {"documentationUrl": "https://docs.airbyte.io/integrations/sources/file", "ConnectionSpecification": {"$schema": "http://json-schema.org/draft-07/schema#", "title": "File Source Spec", "type": "object", "additionalProperties": false, "required": ["dataset_name", "format", "url", "provider"], "properties": {"dataset_name": {"type": "string", "description": "Name of the final table where to replicate this file (should include only letters, numbers dash and underscores)"}, "format": {"type": "string", "enum": ["csv", "json", "jsonl", "excel", "feather", "parquet"], "default": "csv", "description": "File Format of the file to be replicated (Warning: some format may be experimental, please refer to docs)."}, "reader_options": {"type": "string", "description": "This should be a valid JSON string used by each reader/parser to provide additional options and tune its behavior", "examples": ["{}", "{'sep': ' '}"]}, "url": {"type": "string", "description": "URL path to access the file to be replicated"}, "provider": {"type": "object", "description": "Storage Provider or Location of the file(s) to be replicated.", "default": "Public Web", "oneOf": [{"title": "HTTPS: Public Web", "required": ["storage"], "properties": {"storage": {"type": "string", "enum": ["HTTPS"], "default": "HTTPS"}}}, {"title": "GCS: Google Cloud Storage", "required": ["storage"], "properties": {"storage": {"type": "string", "enum": ["GCS"], "default": "GCS"}, "service_account_json": {"type": "string", "description": "In order to access private Buckets stored on Google Cloud, this connector would need a service account json credentials with the proper permissions as described <a href=\"https://cloud.google.com/iam/docs/service-accounts\" target=\"_blank\">here</a>. Please generate the credentials.json file and copy/paste its content to this field (expecting JSON formats). If accessing publicly available data, this field is not necessary."}}}, {"title": "S3: Amazon Web Services", "required": ["storage"], "properties": {"storage": {"type": "string", "enum": ["S3"], "default": "S3"}, "aws_access_key_id": {"type": "string", "description": "In order to access private Buckets stored on AWS S3, this connector would need credentials with the proper permissions. If accessing publicly available data, this field is not necessary."}, "aws_secret_access_key": {"type": "string", "description": "In order to access private Buckets stored on AWS S3, this connector would need credentials with the proper permissions. If accessing publicly available data, this field is not necessary.", "airbyte_secret": true}}}, {"title": "SSH: Secure Shell", "required": ["storage", "user", "host"], "properties": {"storage": {"type": "string", "enum": ["SSH"], "default": "SSH"}, "user": {"type": "string"}, "password": {"type": "string", "airbyte_secret": true}, "host": {"type": "string"}, "port": {"type": "number", "default": 22}}}, {"title": "SCP: Secure copy protocol", "required": ["storage", "user", "host"], "properties": {"storage": {"type": "string", "enum": ["SCP"], "default": "SCP"}, "user": {"type": "string"}, "password": {"type": "string", "airbyte_secret": true}, "host": {"type": "string"}, "port": {"type": "number", "default": 22}}}, {"title": "SFTP: Secure File Transfer Protocol", "required": ["storage", "user", "host"], "properties": {"storage": {"type": "string", "enum": ["SFTP"], "default": "SFTP"}, "user": {"type": "string"}, "password": {"type": "string", "airbyte_secret": true}, "host": {"type": "string"}}}, {"title": "Local Filesystem (limited)", "required": ["storage"], "properties": {"storage": {"type": "string", "description": "WARNING: Note that local storage URL available for read must start with the local mount \"/local/\" at the moment until we implement more advanced docker mounting options...", "enum": ["local"], "default": "local"}}}]}}}}}
*/

/*
Check
docker run --rm -i -v /Users/christianpersson/repos/airbyte:/tmp -v /tmp/airbyte_local:/local  airbyte/source-file check --config /tmp/config.json
{"type": "LOG", "log": {"level": "INFO", "message": "Checking access to file:///local/source/se_users.csv..."}}
{"type": "LOG", "log": {"level": "INFO", "message": "Check succeeded"}}
{"type": "CONNECTION_STATUS", "connectionStatus": {"status": "SUCCEEDED"}}
*/

/*
Discover
docker run --rm -i -v /Users/christianpersson/repos/airbyte:/tmp -v /tmp/airbyte_local:/local  airbyte/source-file discover --config /tmp/config.json
{"type": "LOG", "log": {"level": "INFO", "message": "Discovering schema of test at file:///local/source/se_users.csv..."}}
{"type": "CATALOG", "catalog": {"streams": [{"name": "test", "json_schema": {"$schema": "http://json-schema.org/draft-07/schema#", "type": "object", "properties": {"_id": {"type": "string"}, "gender": {"type": "string"}, "age": {"type": "number"}}}}]}}
*/

/*
Read
docker run --rm -i -v /Users/christianpersson/repos/airbyte:/tmp -v /tmp/airbyte_local:/local  airbyte/source-file read --config /tmp/config.json --catalog /tmp/catalog.json
{"type": "RECORD", "record": {"stream": "test", "data": {"_id": "15839779-3507-4f12-9b5c-a96800aa349c", "gender": "NaN", "age": "NaN"}, "emitted_at": 1621603735000}}
{"type": "RECORD", "record": {"stream": "test", "data": {"_id": "158399dc-b3a2-485e-b916-a84a01269cbf", "gender": "Female", "age": 20.0}, "emitted_at": 1621603735000}}
{"type": "RECORD", "record": {"stream": "test", "data": {"_id": "1583a1a7-8679-4c9a-89ae-a72000a3aa47", "gender": "Female", "age": 75.0}, "emitted_at": 1621603735000}}
...


time docker run --rm -i -v /Users/christianpersson/repos/airbyte:/tmp -v /tmp/airbyte_local:/local airbyte/source-go-example:dev read --config /tmp/config.json --catalog /tmp/catalog.json > /dev/null

go run ./ read --config ../../../airbyte/config.json --catalog ../../../airbyte/catalog.json
*/

/* spec example
{
  "documentationUrl": "https://docs.airbyte.io/integrations/destinations/local-csv",
  "supportsIncremental": true,
  "supported_destination_sync_modes": ["overwrite", "append"],
  "ConnectionSpecification": {
    "$schema": "http://json-schema.org/draft-07/schema#",
    "title": "CSV Destination Spec",
    "type": "object",
    "required": ["destination_path"],
    "additionalProperties": false,
    "properties": {
      "destination_path": {
        "description": "Path to the directory where csv files will be written. The destination uses the local mount \"/local\" and any data files will be placed inside that local mount. For more information check out our <a href=\"https://docs.airbyte.io/integrations/destinations/local-csv\">docs</a>",
        "type": "string",
        "examples": ["/local"]
      }
    }
  }
}

*/

/* config.json
{
  "dataset_name": "test",
  "format":  "csv",
  "url":  "/local/source/se_users.csv",
  "provider": {
    "storage": "local"
  }
}
*/

/* catalog.json
{
  "streams": [
    {
      "stream": {
        "name": "test",
        "json_schema": {
          "$schema": "http://json-schema.org/draft-07/schema#",
          "type": "object",
          "properties": {
            "_id": {
              "type": "string"
            },
            "gender": {
              "type": "string"
            },
            "age": {
              "type": "number"
            }
          }
        },
        "supported_sync_modes": [
          "full_refresh"
        ]
      },
      "sync_mode": "full_refresh",
      "destination_sync_mode": "overwrite"
    }
  ]
}
*/

//	"type": "CATALOG", "catalog": {"streams": [{"name": "test", "json_schema": {"$schema": "http://json-schema.org/draft-07/schema#", "type": "object", "properties": {"_id": {"type": "string"}, "gender": {"type": "string"}, "age": {"type": "number"}}}}]}}

/*
{"host" :"127.0.0.1", "port": "8123", "database": "app1", "username": "app1", "password":"app1"}
*/
