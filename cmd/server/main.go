package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"test/pkg/airbyte"
)

type MyHandler struct {
}

func serveHTTP(w http.ResponseWriter, r *http.Request) error {
	switch r.URL.Path {
	case "/source/check":
		var cnf airbyte.SourceConfig
		if err := json.NewDecoder(r.Body).Decode(&cnf); err != nil {
			return err
		}
		return airbyte.SourceCheck(w, cnf)
	case "/source/discover":
		var cnf airbyte.SourceConfig
		if err := json.NewDecoder(r.Body).Decode(&cnf); err != nil {
			return err
		}
		return airbyte.SourceDiscover(w, cnf)
	case "/source/read":
		var cnf airbyte.ReadConfig
		if err := json.NewDecoder(r.Body).Decode(&cnf); err != nil {
			return err
		}
		return airbyte.SourceRead(w, cnf)
	case "/destination/check":
		var cnf airbyte.DestinationConfig
		if err := json.NewDecoder(r.Body).Decode(&cnf); err != nil {
			return err
		}
		return airbyte.DestinationCheck(w, cnf)
	case "/destination/write":
		var cnf airbyte.WriteConfig
		if rd, err := getReader(w, r, &cnf); err != nil {
			return err
		} else {
			defer rd.Close()
			if err := airbyte.DestinationWrite(rd, cnf); err != nil {
				return err
			}
			return rd.Close()
		}
	default:
		return fmt.Errorf("not found")
	}
}

func (m MyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := serveHTTP(w, r); err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func getReader(w http.ResponseWriter, r *http.Request, v interface{}) (io.ReadCloser, error) {
	reader, err := r.MultipartReader()
	if err != nil {
		return nil, err
	}

	p, err := reader.NextPart()
	if err != nil {
		return nil, err
	} else if p.FormName() != "config" {
		return nil, fmt.Errorf("config is expected")
	}

	if err := json.NewDecoder(p).Decode(v); err != nil {
		p.Close()
		return nil, err
	} else if err := p.Close(); err != nil {
		return nil, err
	}

	p, err = reader.NextPart()
	if err != nil {
		return nil, err
	} else if p.FormName() != "data" {
		return nil, fmt.Errorf("data is expected")
	}

	return p, nil
}

func main() {
	if err := Main(); err != nil {
		panic(err)
	}
}

func Main() error {
	myHandler := &MyHandler{
	}
	fmt.Println("starting server")
	return http.ListenAndServe(":9092", myHandler)
}
