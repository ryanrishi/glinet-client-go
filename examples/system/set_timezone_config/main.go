package main

import (
	"bytes"
	"encoding/json"
	"github.com/ryanrishi/glinet-client-go"
	"log"
	"os"
)

func main() {
	client := glinet.NewClient(os.Getenv("GLINET_USERNAME"), []byte(os.Getenv("GLINET_PASSWORD")))
	params := glinet.SetSystemTimezoneConfigRequest{
		Zonename: "America/New York",
		Timezone: "EST5EDT,M3.2.0,M11.1.0", // TODO is there a better way to derive this information? GL.iNet API seems to require it but it shouldn't be on the caller to pass this in...
	}
	res, err := client.System.SetTimezoneConfig(params)
	if err != nil {
		log.Fatal("Error calling system set_timezone_config: ", err)
	}

	buf, _ := json.Marshal(res)
	log.Printf("system set_timezone_config:\t%s", bytes.NewBuffer(buf))
}
