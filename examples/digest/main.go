package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/ryanrishi/glinet-client-go"
)

func main() {
	c := glinet.NewClient()
	challenge, err := c.Digest.Challenge(context.Background(), "root")
	if err != nil {
		fmt.Println(fmt.Errorf("Error: %v\n", err).Error())
		return
	}

	buf, _ := json.Marshal(challenge)
	fmt.Printf("Challenge: %v\n", bytes.NewBuffer(buf))
}
