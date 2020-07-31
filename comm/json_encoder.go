package comm

import (
	"encoding/json"
	"github.com/anycable/anycable-go/common"
)

type jsonEncoder struct {
}

func (jp jsonEncoder) Unmarshal(data []byte, v interface{}) error{
	return json.Unmarshal(data, &v)
}

func (jp jsonEncoder) MarshalReply(v interface{}) ([]byte, error){
	return json.Marshal(&v)
}

func (jp jsonEncoder) MarshalPing(v interface{}) ([]byte, error){
	return json.Marshal(&v)
}

func (jp jsonEncoder) MarshalDisconnect(v interface{}) ([]byte, error){
	return json.Marshal(&v)
}

func (jp jsonEncoder) MarshalTransmissions(transmissions []string, msg *common.Message) ([][]byte, error){
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
