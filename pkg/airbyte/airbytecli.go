package airbyte

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func Source(w io.Writer) error {
	switch os.Args[1] {
	case "spec":
		return SourceSpec(w)
	case "check":
		var cnf SourceConfig
		if err := readConfig(&cnf); err != nil {
			return err
		}
		return req(join(cnf.Url, "source/check"), cnf, w)
	case "discover":
		var cnf SourceConfig
		if err := readConfig(&cnf); err != nil {
			return err
		}
		return req(join(cnf.Url, "source/discover"), cnf, w)
	case "read":
		var cnf SourceConfig
		var catalog ConfiguredAirbyteCatalog
		var state interface{}
		if err := readConfig(&cnf); err != nil {
			return err
		} else if err := readCatalog(&catalog); err != nil {
			return err
		} else if err := readState(&state); err != nil {
			return err
		}
		return req(join(cnf.Url, "source/read"), ReadConfig{Config: cnf, State: state, Catalog: catalog}, w)
	default:
		return fmt.Errorf("invalid command")
	}
}

func Destination(w io.Writer) error {
	switch os.Args[1] {
	case "spec":
		return DestinationSpec(w)
	case "check":
		var cnf DestinationConfig
		if err := readConfig(&cnf); err != nil {
			return err
		}
		return req(join(cnf.Url, "destination/check"), cnf, w)
	case "write":
		var cnf DestinationConfig
		var catalog ConfiguredAirbyteCatalog
		if err := readConfig(&cnf); err != nil {
			return err
		} else if err := readCatalog(&catalog); err != nil {
			return err
		}
		return ClientSend(join(cnf.Url, "destination/write"), WriteConfig{Config: cnf, Catalog: catalog}, os.Stdin)
	}

	return nil
}

func req(path string, body interface{}, w io.Writer) error {
	var r io.Reader
	if body != nil {
		var buf = bytes.NewBuffer(nil)
		if err := json.NewEncoder(buf).Encode(body); err != nil {
			return err
		}
		r = buf
	}
	return reqReader(path, r, w)
}

func reqReader(path string, r io.Reader, w io.Writer) error {

	resp, err := http.Post(path, "", r)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("invalid status code %d", resp.StatusCode)
	} else if _, err := io.Copy(w, resp.Body); err != nil {
		return err
	}
	return resp.Body.Close()
}

func join(base, path string) string {
	return base + "/" + path
}

func ReadJson(pathOrConfig string, v interface{}) error {
	var data []byte
	var _, err = os.Stat(pathOrConfig)
	if os.IsNotExist(err) {
		data = []byte(pathOrConfig)
	} else if err != nil {
		return err
	} else {
		f, err := os.Open(pathOrConfig)
		if err != nil {
			return err
		}
		defer f.Close()
		data, err = io.ReadAll(f)
		if err != nil {
			return err
		}
	}

	return json.Unmarshal(data, v)
}

func readConfig(v interface{}) error {
	if os.Args[2] != "--config" {
		return fmt.Errorf("expect --config")
	}
	return ReadJson(os.Args[3], v)
}

func readCatalog(v interface{}) error {
	if os.Args[4] != "--catalog" {
		return fmt.Errorf("expect --catalog")
	}
	return ReadJson(os.Args[5], v)
}

func readState(v interface{}) error {
	if len(os.Args) <= 6 {
		return nil
	}
	if os.Args[6] != "--state" {
		return fmt.Errorf("expect --state")
	}
	return ReadJson(os.Args[7], v)
}
