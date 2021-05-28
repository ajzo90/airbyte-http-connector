package main

import (
	"log"
	"os"
	"test/pkg/airbyte"
)

/*
docker run --rm -i <destination-image-name> spec
docker run --rm -i <destination-image-name> check --config <config-file-path>
cat <&0 | docker run --rm -i <destination-image-name> write --config <config-file-path> --catalog <catalog-file-path>
*/

// docker run --rm -i -v /Users/christianpersson/repos/airbyte:/tmp airbyte/destination-csv check --config /tmp/check.json

func main() {
	if err := airbyte.Source(os.Stdout); err != nil {
		log.Fatalln(airbyte.UniqErrStr, err)
	}
}
