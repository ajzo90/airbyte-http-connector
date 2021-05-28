package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"test/pkg/airbyte"
)

func post(url string, in, out interface{}) error {
	var buf = bytes.NewBuffer(nil)
	if err := json.NewEncoder(buf).Encode(in); err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", buf)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		s, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("invalid status code %d %s, %s", resp.StatusCode, resp.Status, string(s))
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

type SourceDestDefReq struct {
	Name                    string `json:"name"`
	DockerRepository        string `json:"dockerRepository"`
	DockerImageTag          string `json:"dockerImageTag"`
	DocumentationUrl        string `json:"documentationUrl"`
	SourceDefinitionId      string `json:"sourceDefinitionId,omitempty"`
	DestinationDefinitionId string `json:"destinationDefinitionId,omitempty"`
}

type SourceDestReq struct {
	Name                    string               `json:"name"`
	SourceDefinitionId      string               `json:"sourceDefinitionId,omitempty"`
	DestinationDefinitionId string               `json:"destinationDefinitionId,omitempty"`
	ConnectionConfiguration airbyte.SourceConfig `json:"connectionConfiguration"`
	WorkSpaceID             string               `json:"workspaceId"`
	SourceId                string               `json:"sourceId,omitempty"`
}

func main() {
	if err := Main(); err != nil {
		log.Fatalln(err)
	}
}

const WorkspaceID = "5ae6b09b-fdec-41af-aaf7-7d94cfc33ef6"

func Main() error {
	const base = "http://localhost:8001/api/v1"
	var destDefReq = SourceDestDefReq{Name: "dest-def", DockerRepository: "airbyte/destination-go-example", DockerImageTag: "dev", DocumentationUrl: "example.com"}
	var sourceDefReq = SourceDestDefReq{Name: "src-def", DockerRepository: "airbyte/source-go-example", DockerImageTag: "dev", DocumentationUrl: "example.com"}

	var destDefResp, srcDefResp SourceDestDefReq

	if err := post(base+"/destination_definitions/create", destDefReq, &destDefResp); err != nil {
		return err
	}
	fmt.Println("Created destination definition", destDefResp)
	if err := post(base+"/source_definitions/create", sourceDefReq, &srcDefResp); err != nil {
		return err
	}
	fmt.Println("Created source definition", srcDefResp)

	var cnf = airbyte.SourceConfig{Url: "http://127.0.0.1:9092", Key: "secret"}
	var destReq = SourceDestReq{Name: "dest", WorkSpaceID: WorkspaceID, DestinationDefinitionId: destDefResp.DestinationDefinitionId, ConnectionConfiguration: cnf}
	var sourceReq = SourceDestReq{Name: "src", WorkSpaceID: WorkspaceID, SourceDefinitionId: srcDefResp.SourceDefinitionId, ConnectionConfiguration: cnf}
	var destResp, srcResp SourceDestReq

	if err := post(base+"/destinations/create", destReq, &destResp); err != nil {
		return err
	}
	fmt.Println("Created destination", destResp)

	if err := post(base+"/sources/create", sourceReq, &srcResp); err != nil {
		return err
	}

	fmt.Println("Created source", srcResp)

	return nil

}
