package airbyte

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/valyala/fastjson"
	"io"
	"log"
	"math"
	"sort"
	"strconv"
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
				"url": {Description: "url", Type: "string", Examples: []string{"http://127.0.0.1:9999"}},
				"key": {Description: "secret", Type: "string", Examples: []string{"secret"}},
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

type strConf struct {
	cols       []string
	extractors []func(o *fastjson.Object, buf []byte) []byte
	wr         *io.PipeWriter
	pkLength   int
	start      int
	stop       int
}

func DestinationWrite(r io.Reader, cnf WriteConfig) error {
	if err := validateDestConfig(cnf.Config); err != nil {
		return err
	}

	var s2c = map[string]strConf{}

	var maxDataCols, maxPK int
	for _, s := range cnf.Catalog.Streams {
		var pks = len(s.PrimaryKey)
		if pks > maxPK {
			maxPK = pks
		}
		var datacols = len(s.Stream.Schema.Properties) - pks
		if datacols > maxDataCols {
			maxDataCols = datacols
		}
	}

	for _, s := range cnf.Catalog.Streams {
		type col struct {
			name    string
			extract func(o *fastjson.Object, buf []byte) []byte
			i       int
		}

		var stream = []byte(s.Stream.Name)
		var cols []col

		for k, v := range s.Stream.Schema.Properties {
			var colDef = col{name: k, i: math.MaxInt64}
			var col = k

			if v.Type == "string" {
				colDef.extract = func(o *fastjson.Object, buf []byte) []byte {
					return o.Get(col).GetStringBytes()
				}
			} else if v.Type == "number" || v.Type == "int" {
				colDef.extract = func(o *fastjson.Object, buf []byte) []byte {
					buf = strconv.AppendFloat(buf[:0], o.Get(col).GetFloat64(), 'f', -1, 64)
					return buf
				}
			} else if v.Type == "object" {
				return fmt.Errorf("object supported soon")
			} else if v.Type == "array" {
				return fmt.Errorf("array supported soon")
			} else if v.Type == "boolean" {
				var one = []byte("1")
				colDef.extract = func(o *fastjson.Object, buf []byte) []byte {
					if o.Get(col).GetBool() {
						return one
					}
					return nil
				}
			} else if v.Type == "null" {
				colDef.extract = func(o *fastjson.Object, buf []byte) []byte {
					return nil
				}
			} else {
				return fmt.Errorf("type %s not supported yet", v.Type)
			}
			cols = append(cols, colDef)
		}

	PkLoop:
		for i, c := range s.PrimaryKey {
			if len(c) != 1 {
				return fmt.Errorf("not support nested PK for now")
			}
			for j := range cols {
				if cols[j].name == c[0] {
					cols[j].i = i
					continue PkLoop
				}
			}
			return fmt.Errorf("not sorted PK, %v <> %v", s.PrimaryKey, cols)
		}

		sort.SliceStable(cols, func(i, j int) bool {
			return cols[i].i < cols[j].i
		})

		switch s.DestinationSyncMode {
		case "append":
		case "overwrite":
		case "append_dedup":
		}

		// not useful here?
		switch s.SyncMode {
		case "incremental":
		case "full_refresh":
		}

		// stream, ....pk , data...., // can subslice [pk,data]

		var cnf = strConf{
			start: maxPK - len(s.PrimaryKey),
			stop:  maxPK + (len(s.Stream.Schema.Properties) - len(s.PrimaryKey)),
		}
		for _, col := range cols {
			cnf.cols = append(cnf.cols, col.name)
			cnf.extractors = append(cnf.extractors, col.extract)
		}

		var dummyVal = []byte("X")
		var dummy = func(o *fastjson.Object, buf []byte) []byte {
			return dummyVal
		}

		for i := 0; i < cnf.start; i++ {
			cnf.extractors = append([]func(o *fastjson.Object, buf []byte) []byte{dummy}, cnf.extractors...)
			cnf.cols = append([]string{"dummyPK"}, cnf.cols...)
		}

		// prepend stream extractor
		cnf.extractors = append([]func(o *fastjson.Object, buf []byte) []byte{func(o *fastjson.Object, buf []byte) []byte {
			return stream
		}}, cnf.extractors...)
		cnf.cols = append([]string{"stream"}, cnf.cols...)

		for i := cnf.stop; i < maxPK+maxDataCols; i++ {
			cnf.extractors = append(cnf.extractors, dummy)
			cnf.cols = append(cnf.cols, "dummyData")
		}

		// shift due to stream
		cnf.start++
		cnf.stop++

		fmt.Println("num extractors", len(cnf.extractors), cnf.start, cnf.stop, cnf.cols)
		s2c[s.Stream.Name] = cnf
	}

	var rd = airbyteReader{sc: bufio.NewScanner(r), s2c: s2c}
	var cells = make([][]byte, maxDataCols+maxPK+1)

	for i := 0; i < 12; i++ {
		if err := rd.Next(cells); err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		log.Println(ToStrings(cells))
	}

	return nil
}

type airbyteReader struct {
	sc  *bufio.Scanner
	p   fastjson.Parser
	s2c map[string]strConf
}

func (a airbyteReader) Next(cells [][]byte) error {
start:
	if !a.sc.Scan() {
		return io.EOF
	}
	res, err := a.p.ParseBytes(a.sc.Bytes())
	if err != nil {
		return err
	}
	var t = res.GetStringBytes("type")
	if string(t) == "RECORD" {
		var record = res.GetObject("record")
		var data = record.Get("data").GetObject()
		var stream = record.Get("stream").GetStringBytes()
		var cnf = a.s2c[string(stream)]

		for i, extr := range cnf.extractors {
			cells[i] = extr(data, cells[i])
		}
		return nil
	}
	goto start
}

func (a airbyteReader) Columns() []string {
	panic("implement me")
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

func ToStrings(strs [][]byte) []string {
	return MapToStrings(len(strs), func(i int) string {
		return string(strs[i])
	})
}

func MapToStrings(N int, mapper func(i int) string) []string {
	var o = make([]string, N)
	for i := range o {
		o[i] = mapper(i)
	}
	return o
}
