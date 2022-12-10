package main

import (
	"bytes"
	"encoding/json"
	"github.com/ryanrishi/glinet-client-go"
	"log"
	"os"
)

func main() {
	params := &glinet.NewClientParams{
		Username: os.Getenv("GLINET_USERNAME"),
		Password: []byte(os.Getenv("GLINET_PASSWORD")),
	}

	client := glinet.NewClient(params)
	res, err := client.System.GetStatus()
	if err != nil {
		log.Fatal("Error calling system get_status: ", err)
	}

	buf, _ := json.Marshal(res)
	log.Printf("system get_status:\t%s", bytes.NewBuffer(buf))
}
