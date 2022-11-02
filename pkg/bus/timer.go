package bus

import (
	"time"

	"github.com/lukerops/pluxy/pkg/commands"
	"github.com/rs/zerolog/log"
)

func (bus *messageBus) AddTimer(timeout time.Duration, cmd commands.Command) {
	bus.mutex.Lock()
	defer bus.mutex.Unlock()

	log.Info().Str("module", "bus.timer").Msg(cmd.String())
	bus.timers = append(bus.timers, timer{
		time: time.Now().Add(timeout),
		cmd:  cmd,
	})
}

func (bus *messageBus) runTimer() {
	for {
		time.Sleep(time.Second)

		now := time.Now()

		var executed []int
		for index, timer := range bus.timers {
			if now.After(timer.time) {
				executed = append(executed, index)
				bus.chRx <- timer.cmd
			}
		}

		for _, index := range executed {
			bus.timers = append(bus.timers[:index], bus.timers[index+1:]...)
		}
	}
}
