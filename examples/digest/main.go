package main

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/ryanrishi/glinet-client-go"
)

func main() {
	c := glinet.NewClientUnauthenticated()
	challenge, err := c.Digest.Challenge("root")
	if err != nil {
		fmt.Println(err)
		return
	}

	buf, _ := json.Marshal(challenge)
	fmt.Printf("Challenge: %v\n", bytes.NewBuffer(buf))
}
