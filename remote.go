package syncthingclient

import (
	"crypto/tls"
	"errors"
	"github.com/golang/protobuf/proto"
	"github.com/syncthing/syncthing/lib/protocol"
	"log"
	"net"
)

type Remote struct {
	conn          net.Conn
	DeviceID      protocol.DeviceID // The device ID of the remote instance
	Host          string
	ClusterConfig *protocol.ClusterConfig
	Callback      Callback
}

//TODO:
// Download Complete
// Upload Complete
// Index Recived
type Callback struct {
	CCrecived chan bool
}

func (r *Remote) Send(data []byte) error {
	_, err := r.conn.Write(data)
	return err
}

func (remote *Remote) listener() {
	reply := make([]byte, 1024)
	for {
		n, err := remote.conn.Read(reply)
		if err != nil {
			log.Fatalf("client: error: %s", err)
		}

		msg := decode(reply[:n])

		log.Printf("New message type = %s\n", msg.GetHeader().Type)

		if msg.GetHeader().Type == 0 {
			rcc := &protocol.ClusterConfig{}
			err = proto.Unmarshal(msg.Message, rcc)
			if err != nil {
				log.Fatal(err)
			}

			remote.ClusterConfig = rcc

			// If we are waiting for the cluster config send that we got it
			if remote.Callback.CCrecived != nil {
				remote.Callback.CCrecived <- true
				break // We have the information we need, stop waiting
			}
		}
	}
}

func (remote *Remote) close() {
	remote.conn.Close()
}

func (remote *Remote) connect(cert tls.Certificate) error {
	config := tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}
	conn, err := tls.Dial("tcp", remote.Host, &config)
	if err != nil {
		return errors.New("Client: dial: " + err.Error())
	}

	log.Println("Connected to: ", conn.RemoteAddr())

	// Send hello
	hello := &protocol.Hello{
		DeviceName:    "device",
		ClientName:    "LiteClient",
		ClientVersion: "v0.0.1",
	}

	rHello, err := protocol.ExchangeHello(conn, hello)
	if err != nil {
		return err
	}

	log.Println("Got hello from " + rHello.DeviceName + " (" + rHello.ClientName + " " + rHello.ClientVersion + ")")

	state := conn.ConnectionState()
	if len(state.PeerCertificates) != 1 {
		return errors.New("To many certificates")
	}

	remoteID := protocol.NewDeviceID(state.PeerCertificates[0].Raw)
	log.Printf("Remote ID = %s", remoteID)

	if remoteID != remote.DeviceID {
		return errors.New("Device id did not match with certificate")
	}

	if !state.HandshakeComplete {
		return errors.New("Handshake not Complete")
	}

	remote.conn = conn

	go remote.listener()

	return nil
}

// SetHost sets the remote host name host:port format
func (r *Remote) SetHost(Host string) {
	r.Host = Host
}
