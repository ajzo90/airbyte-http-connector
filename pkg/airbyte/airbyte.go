package airbyte

const UniqErrStr = "APPERROR" // simplify search in logs.. :)

type AirbyteCatalogWrapper struct {
	Type    string         `json:"type"`
	Catalog AirbyteCatalog `json:"catalog"`
}

type AirbyteStream struct {
	Name                    string     `json:"name"`
	Namespace               string     `json:"namespace"`
	Schema                  JSONSchema `json:"json_schema"`
	SupportedSyncModes      []string   `json:"supported_sync_modes,omitempty"` //["full_refresh", "incremental"]
	SourceDefinedCursor     bool       `json:"source_defined_cursor,omitempty"`
	DefaultCursorField      []string   `json:"default_cursor_field,omitempty"`
	SourceDefinedPrimaryKey [][]string `json:"source_defined_primary_key,omitempty"` //[["resultId"], ["resultIdx"], ["i64"]]
}

type JSONSchema struct {
	Schema     string              `json:"$schema"`
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties"`
}

type AirbyteCatalog struct {
	Streams []AirbyteStream `json:"streams"`
}

type ConfiguredAirbyteCatalog struct {
	Streams []ConfiguredAirbyteStream `json:"streams"`
}

// required: stream, sync_mode, destination_sync_mode
type ConfiguredAirbyteStream struct {
	Stream              AirbyteStream `json:"stream"`
	SyncMode            string        `json:"sync_mode"`             //full_refresh [full_refresh,incremental]
	CursorField         []string      `json:"cursor_field"`          //Path to the field that will be used to determine if a record is new or modified since the last sync. This field is REQUIRED if `sync_mode` is `incremental`. Otherwise it is ignored.
	DestinationSyncMode string        `json:"destination_sync_mode"` //append [append, overwrite, append_dedup, upsert_dedup]
	PrimaryKey          [][]string    `json:"primary_key"`           // Paths to the fields that will be used as primary key. This field is REQUIRED if `destination_sync_mode` is `*_dedup`. Otherwise it is ignored.
}

type ConnectionStatusWrap struct {
	Type             string           `json:"type"`
	ConnectionStatus ConnectionStatus `json:"connectionStatus"`
}

type ConnectionStatus struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

type ConnectorSpecificationWrap struct {
	Type string                 `json:"type"`
	Spec ConnectorSpecification `json:"spec"`
}

type ConnectorSpecification struct {
	DocumentationUrl              string                  `json:"documentationUrl"`
	SupportsIncremental           bool                    `json:"supportsIncremental,omitempty"`
	SupportedDestinationSyncModes []string                `json:"supported_destination_sync_modes,omitempty"`
	ConnectionSpecification       ConnectionSpecification `json:"connectionSpecification"`
}

type ConnectionSpecification struct {
	Schema               string              `json:"$schema"`
	Title                string              `json:"title"`
	Type                 string              `json:"type"`
	Required             []string            `json:"required"`
	AdditionalProperties bool                `json:"additionalProperties"`
	Properties           map[string]Property `json:"properties"`
}

type Property struct {
	Description string   `json:"description,omitempty"`
	Type        string   `json:"type"`
	Examples    []string `json:"examples,omitempty"`
}

type AirbyteLogMessage struct {
	Level   string `json:"level"`
	Message string `json:"message"`
}

type AirbyteLogMessageWrap struct {
	Type string            `json:"type"`
	Log  AirbyteLogMessage `json:"log"`
}
