package bus

import (
	"sync"

	"github.com/lukerops/pluxy/pkg/commands"
)

type Handler interface {
	Run(tx chan<- commands.Command, rx <-chan commands.Command)
}

type handlerInfo struct {
	name    commands.CommandHandler
	handler Handler
	chTx    chan commands.Command
}

type messageBus struct {
	handlerInfo map[commands.CommandHandler]*handlerInfo
	chRx        chan commands.Command
	mutex       sync.RWMutex
}

var MessageBus *messageBus

func NewMessageBus() {
	MessageBus = &messageBus{
		handlerInfo: make(map[commands.CommandHandler]*handlerInfo),
		chRx:        make(chan commands.Command, 10),
	}
}
