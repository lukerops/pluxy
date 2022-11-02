package bus

import "github.com/lukerops/pluxy/pkg/commands"

func (bus *messageBus) Register(name commands.CommandHandler, handler Handler) {
	bus.mutex.Lock()
	defer bus.mutex.Unlock()

	bus.handlerInfo[name] = &handlerInfo{
		name:    name,
		handler: handler,
		chTx:    make(chan commands.Command, 5),
	}
}
