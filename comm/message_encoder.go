package comm

import (
	"github.com/anycable/anycable-go/common"
)

type MessageEncoder interface {
	MarshalReply(v interface{}) ([]byte, error)
	MarshalPing(v interface{}) ([]byte, error)
	MarshalDisconnect(v interface{}) ([]byte, error)
	MarshalTransmissions(transmissions []string, msg *common.Message) ([][]byte, error)
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
