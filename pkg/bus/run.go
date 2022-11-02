package bus

import (
	"github.com/lukerops/pluxy/pkg/commands"
	"github.com/rs/zerolog/log"
)

func (bus *messageBus) Run() {
	// o Tx do bus é o Rx do handler
	for _, handlerInfo := range bus.handlerInfo {
		handlerInfo.handler.Run(bus.chRx, handlerInfo.chTx)
	}

	go bus.runTimer()
	go bus.run()
}

func (bus *messageBus) run() {
	logger := log.With().Str("module", "bus").Logger()

	for {
		cmd := <-bus.chRx

		bus.mutex.RLock()
		logger.Info().Msg(cmd.String())

		if cmd.Cmd == commands.CommandStop {
			for _, handlerInfo := range bus.handlerInfo {
				handlerInfo.chTx <- cmd
			}
			return
		}

		sendTo := cmd.To
		if cmd.IsResponse() {
			sendTo = cmd.From
		}

		bus.handlerInfo[sendTo].chTx <- cmd
		bus.mutex.RUnlock()
	}
}
