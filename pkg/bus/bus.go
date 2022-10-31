package bus

import (
	"sync"

	"github.com/lukerops/pluxy/pkg/commands"
)

type messageBus struct {
	chTx  map[string]chan commands.Command
	chRx  chan commands.Command
	mutex sync.RWMutex
}

var MessageBus *messageBus

func NewMessageBus() {
    MessageBus = &messageBus{
        chTx: make(map[string]chan commands.Command),
		chRx: make(chan commands.Command, 10),
	}
}

