package main

import (
	"encoding/json"
	"github.com/ryanrishi/glinet-client-go"
	"log"
	"os"
)

type system struct {
	LanIP string `json:"lan_ip"`
}

type systemStatus struct {
	System system `json:"system"`
}

func main() {
	params := &glinet.NewClientParams{
		Username: os.Getenv("GLINET_USERNAME"),
		Password: []byte(os.Getenv("GLINET_PASSWORD")),
	}

	client := glinet.NewClient(params)
	var res systemStatus
	reqParams := []string{client.Sid, "system", "get_status"}
	err := client.CallWithStringSlice("call", reqParams, &res)
	if err != nil {
		log.Fatal("Error calling system get_status: ", err)
	}

	buf, _ := json.Marshal(res)
	log.Print(buf)
}
