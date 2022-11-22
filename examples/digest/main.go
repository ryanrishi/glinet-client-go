package main

import (
	"context"
	"fmt"
	"github.com/ryanrishi/glinet-client-go"
)

func main() {
	c := glinet.NewClient(nil)
	challenge, _, err := c.Digest.Challenge(context.Background(), "admin")
	if err != nil {
		fmt.Println(fmt.Errorf("Error: %v\n", err).Error())
		return
	}

	fmt.Printf("Challenge: %v\n", challenge)
}
