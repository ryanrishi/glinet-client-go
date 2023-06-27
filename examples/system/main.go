package main

import (
	"bytes"
	"encoding/json"
	"log"
	"os"

	"github.com/ryanrishi/glinet-client-go"
)

func main() {
	client := glinet.NewClient(os.Getenv("GLINET_USERNAME"), []byte(os.Getenv("GLINET_PASSWORD")))
	res, err := client.System.GetStatus()
	if err != nil {
		log.Fatal("Error calling system get_status: ", err)
	}

	buf, _ := json.Marshal(res)
	log.Printf("system get_status:\t%s", bytes.NewBuffer(buf))
}
