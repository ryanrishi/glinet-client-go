package main

import "github.com/ryanrishi/glinet-client-go"

func main() {
	c := glinet.NewClient(nil)
	c.Digest.Login()
}
