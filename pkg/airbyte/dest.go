package airbyte

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/valyala/fastjson"
	"io"
)

type DestinationConfig struct {
	Url string `json:"url"`
	Key string `json:"key"`
}

func DestinationSpec(w io.Writer) error {
	return json.NewEncoder(w).Encode(ConnectorSpecificationWrap{Type: "SPEC", Spec: ConnectorSpecification{
		DocumentationUrl:              "https://docs.airbyte.io/integrations/destinations/mssql",
		SupportedDestinationSyncModes: []string{"append", "overwrite", "append_dedup"}, //upsert_dedup
		ConnectionSpecification: ConnectionSpecification{
			Schema:               "http://json-schema.org/draft-07/schema#",
			Title:                "Airbyte HTTP Destination Spec",
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

func validateDestConfig(cnf DestinationConfig) error {
	if cnf.Key != "secret" {
		return fmt.Errorf("invalid key")
	}
	return nil
}

func DestinationCheck(w io.Writer, cnf DestinationConfig) error {
	var ret = ConnectionStatusWrap{Type: "CONNECTION_STATUS", ConnectionStatus: ConnectionStatus{Status: "SUCCEEDED"}}
	if err := validateDestConfig(cnf); err != nil {
		ret = ConnectionStatusWrap{Type: "CONNECTION_STATUS", ConnectionStatus: ConnectionStatus{Status: "FAILED", Message: err.Error()}}
	}
	return json.NewEncoder(w).Encode(ret)
}

type WriteConfig struct {
	Config  DestinationConfig        `json:"config"`
	Catalog ConfiguredAirbyteCatalog `json:"catalog"`
}

func DestinationWrite(r io.Reader, cnf WriteConfig) error {
	if err := validateDestConfig(cnf.Config); err != nil {
		return err
	}
	for _, s := range cnf.Catalog.Streams {
		switch s.DestinationSyncMode {
		case "append":
		case "overwrite":
		case "append_dedup":
		}

		switch s.SyncMode {
		case "incremental":
		case "full_refresh":
		}
	}

	var sc = bufio.NewScanner(r)
	var p fastjson.Parser
	var records int

	for ; sc.Scan(); {
		var b = sc.Bytes()
		res, err := p.ParseBytes(b)
		if err != nil {
			return err
		}
		var t = res.GetStringBytes("type")
		if bytes.Equal(t, []byte("RECORD")) {
			// todo: do something meaningful with the records...
			if records == 0 {
				fmt.Println(string(b))
			}
			records++
		} else if bytes.Equal(t, []byte("STATE")) {
			fmt.Println("state")
			fmt.Println(string(b))
		}
	}
	fmt.Println("records", records)

	return nil
}

// curl http://localhost:8001/api/v1/destination_definitions/create -H 'Content-Type:application/json' -d '{"name": "dest","dockerRepository": "airbyte/destination-go-example","dockerImageTag": "dev","documentationUrl": "http://example.com"}'
// curl http://localhost:8001/api/v1/source_definitions/create -H 'Content-Type:application/json' -d '{"name": "src","dockerRepository": "airbyte/source-go-example","dockerImageTag": "dev","documentationUrl": "http://example.com"}'

// curl http://localhost:8001/api/v1/source_definitions/create -H 'Content-Type:application/json' -d '{"name": "src","dockerRepository": "ajzo90/airbyte-source-go-example","dockerImageTag": "dev","documentationUrl": "http://example.com"}'
// curl http://localhost:8001/api/v1/destination_definitions/create -H 'Content-Type:application/json' -d '{"name": "dest2","dockerRepository": "ajzo90/airbyte-destination-go-example","dockerImageTag": "dev","documentationUrl": "http://example.com"}'

/*
spec() -> ConnectorSpecification
check(Config) -> AirbyteConnectionStatus
write(Config, AirbyteCatalog, Stream<AirbyteMessage>(stdin)) -> void
*/

/*
docker run --rm -i <destination-image-name> spec
docker run --rm -i <destination-image-name> check --config <config-file-path>
cat <&0 | docker run --rm -i <destination-image-name> write --config <config-file-path> --catalog <catalog-file-path>
*/
