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
	res, err := client.AdGuard.GetAdGuardConfig()
	if err != nil {
		log.Fatal("Error getting AdGuard config: ", err)
	}

	buf, _ := json.Marshal(res)
	log.Printf("AdGuard config:\t%s", bytes.NewBuffer(buf))

	err = client.AdGuard.SetAdGuardConfig(!res.Enabled)
	if err != nil {
		log.Fatal("Error setting AdGuard config: ", err)
	}
}
