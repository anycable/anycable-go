package comm

import (
	"encoding/json"
	"github.com/anycable/anycable-go/common"
)

type jsonEncoder struct {
}

func (jp jsonEncoder) Unmarshal(data []byte, message **common.Message) error {
	return json.Unmarshal(data, message)
}

func (jp jsonEncoder) MarshalIsBinary() bool {
	return false
}

func (jp jsonEncoder) MarshalReply(message *common.Reply) ([]byte, error) {
	return json.Marshal(&message)
}

func (jp jsonEncoder) MarshalPing(message *common.PingMessage) ([]byte, error) {
	return json.Marshal(&message)
}

func (jp jsonEncoder) MarshalDisconnect(message *common.DisconnectMessage) ([]byte, error) {
	return json.Marshal(&message)
}

func (jp jsonEncoder) MarshalTransmissions(transmissions []string, message *common.Message) ([][]byte, error) {
	return marshalTransmissions(transmissions)
}

func (jp jsonEncoder) MarshalAuthenticateTransmissions(transmissions []string) ([][]byte, error) {
	return marshalTransmissions(transmissions)
}

func marshalTransmissions(transmissions []string) ([][]byte, error) {
	var transmissionBytes [][]byte
	for _, transmission := range transmissions {
		transmissionBytes = append(transmissionBytes, []byte(transmission))
	}

	return transmissionBytes, nil
}
