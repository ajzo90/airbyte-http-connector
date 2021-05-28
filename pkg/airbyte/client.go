package airbyte

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

// curl -F 'file_field=@../../../platform-data/rusta/se_trans2.csv' http://127.0.0.1:8082
func ClientSend(url string, cnf interface{}, in io.Reader) error {
	r, w := io.Pipe()
	m := multipart.NewWriter(w)
	
	go func() {

		err := func() error {
			wr, err := m.CreateFormField("config")
			if err != nil {
				return err
			} else if err := json.NewEncoder(wr).Encode(cnf); err != nil {
				return err
			}

			part, err := m.CreateFormFile("data", "foo.txt")
			if err != nil {
				return err
			} else if _, err = io.Copy(part, in); err != nil {
				return err
			}
			return nil
		}()

		m.Close()
		w.CloseWithError(err)
	}()

	res, err := http.Post(url, m.FormDataContentType(), r)
	if err != nil {
		return err
	} else if res.StatusCode != 200 {
		return fmt.Errorf("bad status %d", res.StatusCode)
	}
	return nil
}
