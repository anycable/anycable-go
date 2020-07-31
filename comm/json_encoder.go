package comm

import (
	"encoding/json"
	"github.com/anycable/anycable-go/common"
	"github.com/anycable/anycable-go/node"
)

type jsonEncoder struct {
}

func (jp jsonEncoder) Unmarshal(data []byte, v interface{}) error{
	return json.Unmarshal(data, &v)
}

func (jp jsonEncoder) MarshalReply(message *node.Reply) ([]byte, error){
	return json.Marshal(&message)
}

func (jp jsonEncoder) MarshalPing(message *node.PingMessage) ([]byte, error){
	return json.Marshal(&message)
}

func (jp jsonEncoder) MarshalDisconnect(message *node.DisconnectMessage) ([]byte, error){
	return json.Marshal(&message)
}

func (jp jsonEncoder) MarshalTransmissions(transmissions []string, message *common.Message) ([][]byte, error){
	var transmissionBytes [][]byte
	for _, transmission := range transmissions{
		transmissionBytes = append(transmissionBytes, []byte(transmission))
	}

	return transmissionBytes, nil
}

func (jp jsonEncoder) MarshalAuthenticateTransmissions(transmissions []string) ([][]byte, error){
	var transmissionBytes [][]byte
	for _, transmission := range transmissions{
		transmissionBytes = append(transmissionBytes, []byte(transmission))
	}

	return transmissionBytes, nil
}
