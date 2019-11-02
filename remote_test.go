package syncthingclient

import (
	"bytes"
	"github.com/golang/protobuf/proto"
	"github.com/syncthing/syncthing/lib/protocol"
	"testing"
)

func TestGetMessage(t *testing.T) {
	data := make([]byte, 5000)
	data[4999] = 1
	input, _ := MakeResponse(0, data)

	r := bytes.NewReader(input)

	msg, err := getMessage(r)
	if err != nil {
		t.Error("Got error " + err.Error())
	}

	response := &protocol.Response{}
	err = proto.Unmarshal(msg.Message, response)
	if err != nil {
		t.Error("Got error " + err.Error())
	}

	if len(response.Data) != 5000 {
		t.Error("Did not get expected data")
	}

}
