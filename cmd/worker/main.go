package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/valyala/fastjson"
	"log"
	"os"
	"os/exec"
)

func docker(image string, command ...string) *exec.Cmd {
	//const DATA_MOUNT_DESTINATION = "/data"
	//const LOCAL_MOUNT_DESTINATION = "/local"

	//var localMountSource = ""
	//var v1 = fmt.Sprintf("%s:%s", workspaceMountSource, DATA_MOUNT_DESTINATION)
	//var v2 = fmt.Sprintf("%s:%s", localMountSource, LOCAL_MOUNT_DESTINATION)

	var args = []string{"run", "--rm", "--init", "-i", image}
	args = append(args, command...)
	fmt.Println(args)
	fmt.Println(os.Getenv("PATH"))
	var c = exec.Command("docker", args...)
	return c
}

func Main() error {
	var srcImage = os.Args[1]
	var srcConfig = os.Args[2]
	var srcCatalog = os.Args[3]
	var srcState = os.Args[4]

	var destImage = os.Args[5]
	var destConfig = os.Args[6]
	var destCatalog = os.Args[7]

	var src = docker(srcImage, "read", "--config", srcConfig, "--catalog", srcCatalog, "--state", srcState)
	srcOutStream, err := src.StdoutPipe()
	if err != nil {
		return err
	}
	defer srcOutStream.Close()

	var dest = docker(destImage, "write", "--config", destConfig, "--catalog", destCatalog)
	destInStream, err := dest.StdinPipe()
	if err != nil {
		return err
	}
	defer destInStream.Close()

	if err := src.Start(); err != nil {
		return err
	} else if err := dest.Start(); err != nil {
		return err
	}

	var sc = bufio.NewScanner(srcOutStream)
	var p fastjson.Parser

	var records int

	for ; sc.Scan(); {
		var b = sc.Bytes()
		res, err := p.ParseBytes(b)
		if err != nil {
			return err
		}
		var t = string(res.GetStringBytes("type"))
		switch t {
		case "RECORD":
			if _, err := destInStream.Write(b); err != nil {
				return err
			}
			records++
		case "STATE":
		case "LOG":
		}
	}

	if err := src.Wait(); err != nil {
		return err
	} else if err := dest.Wait(); err != nil {
		return err
	}

	return json.NewEncoder(os.Stdout).Encode(struct {
		Records int
	}{Records: records})

}

func main() {
	if err := Main(); err != nil {
		fmt.Println(err)
		log.Fatalln(err)
	}
}
