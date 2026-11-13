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
	ctx := context.Background()
	client, err := glinet.NewClient(os.Getenv("GLINET_USERNAME"), []byte(os.Getenv("GLINET_PASSWORD")))
	if err != nil {
		log.Fatal("Error creating client: ", err)
	}

	res, err := client.AdGuard.GetAdGuardConfig(ctx)
	if err != nil {
		log.Fatal("Error getting AdGuard config: ", err)
	}

	buf, _ := json.Marshal(res)
	log.Printf("AdGuard config:\t%s", bytes.NewBuffer(buf))

	err = client.AdGuard.SetAdGuardConfig(ctx, &glinet.SetAdGuardConfigRequest{Enabled: !res.Enabled, DNSEnabled: res.DNSEnabled})
	if err != nil {
		log.Fatal("Error setting AdGuard config: ", err)
	}
}
