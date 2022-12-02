package main

import (
	"bytes"
	"encoding/json"
	"github.com/ryanrishi/glinet-client-go"
	"log"
	"os"
)

func main() {
	client := glinet.NewClient()
	login, err := client.Digest.Login(os.Getenv("GLINET_USERNAME"), []byte(os.Getenv("GLINET_PASSWORD")))
	if err != nil {
		log.Fatal("Error running login example: ", err)
	}

	buf, _ := json.Marshal(login)
	log.Printf("Login: %s", bytes.NewBuffer(buf))
}
