package comm

import (
	"github.com/anycable/anycable-go/common"
	"github.com/anycable/anycable-go/node"
)

type MessageEncoder interface {
	MarshalReply(message *node.Reply) ([]byte, error)
	MarshalPing(message *node.PingMessage) ([]byte, error)
	MarshalDisconnect(message *node.DisconnectMessage) ([]byte, error)
	MarshalTransmissions(transmissions []string, message *common.Message) ([][]byte, error)
	MarshalAuthenticateTransmissions(transmissions []string) ([][]byte, error)
	Unmarshal(data []byte, v interface{}) error
}

var messageEncoder MessageEncoder

func GetMessageEncoder() MessageEncoder {
	if messageEncoder == nil {
		messageEncoder = jsonEncoder{}
	}
	return messageEncoder
}
