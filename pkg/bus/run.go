package bus

import (
	"github.com/lukerops/pluxy/pkg/commands"
	"github.com/rs/zerolog/log"
)

func (bus *messageBus) Run() {
	logger := log.With().Str("module", "bus").Logger()

	go func() {
		for {
			cmd := <-bus.chRx

			bus.mutex.RLock()
			logger.Info().Msg(cmd.String())

			if cmd.Cmd == commands.CommandStop {
				for _, ch := range bus.chTx {
					ch <- cmd
				}
				return
			}

			sendTo := cmd.To
			if cmd.IsResponse() {
				sendTo = cmd.From
			}

			bus.chTx[sendTo] <- cmd
			bus.mutex.RUnlock()
		}
	}()
}
