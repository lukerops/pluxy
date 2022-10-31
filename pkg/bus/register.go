package bus

import "github.com/lukerops/pluxy/pkg/commands"

type Handler interface {
    Run(tx chan<- commands.Command, rx <-chan commands.Command)
}

func (bus *messageBus) Register(name string, handler Handler) {
    bus.mutex.Lock()
    defer bus.mutex.Unlock()

    bus.chTx[name] = make(chan commands.Command, 5)

    // o Tx do bus é o Rx do handler
    handler.Run(bus.chRx, bus.chTx[name])
}
