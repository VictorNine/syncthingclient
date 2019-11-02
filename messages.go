package syncthingclient

import (
	"encoding/binary"
	"github.com/golang/protobuf/proto"
	"github.com/syncthing/syncthing/lib/protocol"
	"log"
)

type Message struct {
	HeaderLength  uint16
	Header        []byte
	MessageLength uint32
	Message       []byte
}

// Returns Message, if the message is not complete expected length is also returned
func decode(data []byte) (Message, int) {
	msg := Message{}

	msg.HeaderLength = binary.BigEndian.Uint16(data[0:2])

	if msg.HeaderLength == 0 {
		msg.Header = []byte{}
	} else {
		msg.Header = data[2 : msg.HeaderLength+2]
	}

	MsgStart := 2 + msg.HeaderLength
	msg.MessageLength = binary.BigEndian.Uint32(data[MsgStart : MsgStart+4])
	msg.Message = data[MsgStart+4:]

	if len(msg.Message) != int(msg.MessageLength) {
		return msg, int(msg.MessageLength) + 4 + int(msg.HeaderLength) + 2
	}

	return msg, 0
}

func (msg *Message) GetHeader() *protocol.Header {
	if msg.HeaderLength == 0 {
		return &protocol.Header{
			Type:        0,
			Compression: 0,
		}
	}

	header := &protocol.Header{}
	err := proto.Unmarshal(msg.Header, header)
	if err != nil {
		log.Fatal("GetHeader() " + err.Error())
	}

	return header
}

func makeMessage(MsgType int, Message []byte) ([]byte, error) {
	header := &protocol.Header{
		Type:        protocol.MessageType(MsgType),
		Compression: 0,
	}

	hb, err := proto.Marshal(header)
	if err != nil {
		return nil, err
	}

	headerLength := len(hb)

	msgBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(msgBytes, uint16(headerLength))

	msgBytes = append(msgBytes, hb...)

	// Header done

	msgLength := len(Message)
	ml := make([]byte, 4)
	binary.BigEndian.PutUint32(ml, uint32(msgLength))

	msgBytes = append(msgBytes, ml...)
	msgBytes = append(msgBytes, Message...)

	return msgBytes, nil
}

func (remote *Remote) MakeClusterConfig(folderID string) ([]byte, error) {
	cc := &protocol.ClusterConfig{}

	if len(folderID) > 0 {
		device := []protocol.Device{
			protocol.Device{
				ID:          remote.DeviceID,
				Name:        "",
				Addresses:   []string{"tcp://" + remote.Host},
				Compression: 1,
				CertName:    "",
				//			MaxSequence: 1,
				Introducer: false,
				//			IndexId:                  11383273935130040216,
				//			SkipIntroductionRemovals: false,
			},
		}

		cc.Folders = []protocol.Folder{
			protocol.Folder{
				ID:                 folderID,
				Label:              "",
				ReadOnly:           false,
				IgnorePermissions:  false,
				IgnoreDelete:       false,
				DisableTempIndexes: false,
				Paused:             false,
				Devices:            device,
			},
		}
	}

	data, err := proto.Marshal(cc)
	if err != nil {
		return nil, err
	}

	m, err := makeMessage(0, data)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func MakeRequst(ID int32, rootFolder string, name string, offset int64, size int32, hash []byte) ([]byte, error) {
	req := &protocol.Request{
		ID:            ID,
		Folder:        rootFolder,
		Name:          name, // Full namepath relative to the folder root
		Offset:        offset,
		Size:          size,
		Hash:          hash,
		FromTemporary: false,
	}

	data, err := proto.Marshal(req)
	if err != nil {
		return nil, err
	}

	m, err := makeMessage(3, data)
	if err != nil {
		return nil, err
	}

	return m, nil
}

// Rigth now only used for testing
func MakeResponse(id int32, data []byte) ([]byte, error) {
	resp := &protocol.Response{
		ID:   id,
		Data: []byte(data),
		Code: 0,
	}

	data, err := proto.Marshal(resp)
	if err != nil {
		return nil, err
	}

	m, err := makeMessage(4, data)
	if err != nil {
		return nil, err
	}

	return m, nil
}
