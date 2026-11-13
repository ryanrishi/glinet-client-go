package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/ryanrishi/glinet-client-go"
)

func main() {
	client, err := glinet.NewClient(os.Getenv("GLINET_USERNAME"), []byte(os.Getenv("GLINET_PASSWORD")))
	if err != nil {
		log.Fatal("Error creating client: ", err)
	}

	res, err := client.System.GetStatus(context.Background())
	if err != nil {
		log.Fatal("Error calling system get_status: ", err)
	}

	buf, _ := json.Marshal(res)
	log.Printf("system get_status:\t%s", bytes.NewBuffer(buf))
}
