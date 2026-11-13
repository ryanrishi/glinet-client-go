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
	client := glinet.NewClientUnauthenticated()
	login, err := client.Digest.Login(context.Background(), os.Getenv("GLINET_USERNAME"), []byte(os.Getenv("GLINET_PASSWORD")))
	if err != nil {
		log.Fatal("Error running login example: ", err)
	}

	buf, _ := json.Marshal(login)
	log.Printf("Login: %s", bytes.NewBuffer(buf))
}
