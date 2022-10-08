// pub represents a Publisher Endpoint component,
// the component responsible for accepting broadcasts and commands
// from outside and trigger the publish-record(via broker)-distribute(to clients) pipeline.
package pub

import "github.com/anycable/anycable-go/common"

type Handler interface {
	PublishStreamMessage(msg *common.StreamMessage)
	PublishCommand(cmd interface{})
}
