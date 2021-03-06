package syncthingclient

import (
	"crypto/tls"
	"errors"
	"github.com/golang/protobuf/proto"
	"github.com/syncthing/syncthing/lib/protocol"
	"io"
	"log"
	"net"
	"strings"
)

type Remote struct {
	conn          net.Conn
	DeviceID      protocol.DeviceID // The device ID of the remote instance
	Host          string
	done          bool // Operation is done, connection no longer needed
	ClusterConfig *protocol.ClusterConfig
	Index         *protocol.Index
	Callback      Callback
}

//TODO:
// Upload Complete
type Callback struct {
	CCrecived       chan bool
	IndexRecived    chan bool
	ResponseRecived chan *protocol.Response
}

func (r *Remote) Send(data []byte) error {
	_, err := r.conn.Write(data)
	return err
}

// get message from reader
func getMessage(reader io.Reader) (Message, error) {
	reply := make([]byte, 1024)

	n, err := reader.Read(reply)
	if err != nil {
		return Message{}, err
	}

	msg, left := decode(reply[:n])
	if left > 0 { // Did not get full message
		left = left - n
		buffer := make([]byte, 0) // TODO: Allocate the full buffer in stead
		buffer = append(buffer, reply[:n]...)

		for {
			re := io.LimitReader(reader, int64(len(reply)))

			if len(reply) > left {
				re = io.LimitReader(reader, int64(left))
			}

			n, err := re.Read(reply)
			if err != nil {
				return msg, err
			}
			buffer = append(buffer, reply[:n]...)

			left = left - n
			if left == 0 {
				break
			}
		}
		msg, _ = decode(buffer)
	}

	return msg, nil
}

func (remote *Remote) listener() {
	for !remote.done {
		msg, err := getMessage(remote.conn)

		if err != nil {
			// If we are done this error is ok, and we stop the loop
			if strings.Contains(err.Error(), "use of closed network connection") && remote.done {
				break
			}
			log.Fatalf("client: error: %s", err)
		}

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
			}
		}

		// INDEX
		if msg.GetHeader().Type == 1 {
			index := &protocol.Index{}
			err := proto.Unmarshal(msg.Message, index)
			if err != nil {
				log.Fatal(err)
			}

			remote.Index = index

			// If we are waiting for the index send that we got it
			if remote.Callback.IndexRecived != nil {
				remote.Callback.IndexRecived <- true
			}
		}

		// RESPONSE
		if msg.GetHeader().Type == 4 {
			response := &protocol.Response{}
			err := proto.Unmarshal(msg.Message, response)
			if err != nil {
				log.Fatal("Listen() RESPONSE " + err.Error())
			}

			if response.Code != 0 {
				log.Fatalf("Response error: %v\n", response.Code)
			}

			remote.Callback.ResponseRecived <- response
		}
	}
}

func (remote *Remote) close() {
	remote.done = true // We are done and the for listener can be stopped
	remote.conn.Close()
}

func Handshake(conn *tls.Conn, hello *protocol.Hello, exptectedID protocol.DeviceID) error {
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

	if remoteID != exptectedID {
		return errors.New("Device id did not match with certificate")
	}

	if !state.HandshakeComplete {
		return errors.New("Handshake not Complete")
	}

	return nil
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

	err = Handshake(conn, hello, remote.DeviceID)
	if err != nil {
		return err
	}

	remote.conn = conn

	go remote.listener()

	return nil
}

// SetHost sets the remote host name host:port format
func (r *Remote) SetHost(Host string) {
	r.Host = Host
}
