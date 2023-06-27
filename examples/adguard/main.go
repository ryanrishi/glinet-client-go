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
