package main

import (
	"github.com/VictorNine/syncthingclient"
	"log"
)

func main() {
	client := syncthingclient.Client{}
	err := client.LoadCertificate("certs/cert.pem", "certs/key.pem")
	if err != nil {
		log.Fatal(err)
	}

	err = client.AddRemote(Add the remote devideID here)
	if err != nil {
		log.Fatal(err)
	}

	client.Remote.SetHost("localhost:22000")

	// File Download
	err = client.Download("itgra-hbeuo/test.jpg", "/tmp/test.jpg")
	if err != nil {
		log.Fatal(err)
	}
}
