package comm

import (
	"github.com/anycable/anycable-go/common"
)

type MessageEncoder interface {
	Unmarshal(data []byte, v interface{}) error
	MarshalReply(message *common.Reply) ([]byte, error)
	MarshalPing(message *common.PingMessage) ([]byte, error)
	MarshalDisconnect(message *common.DisconnectMessage) ([]byte, error)
	MarshalTransmissions(transmissions []string, message *common.Message) ([][]byte, error)
	MarshalAuthenticateTransmissions(transmissions []string) ([][]byte, error)
}

func GetMessageEncoder() MessageEncoder {
	return jsonEncoder{}
}
