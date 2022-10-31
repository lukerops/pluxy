package commands

import (
	"fmt"
	"strings"
)

type Pluto Command

func NewPlutoRequest(from string) Pluto {
	return Pluto{
		From: strings.ToUpper(from),
		To:   ToPluto,
	}
}

func NewPlutoResponse(from, response string) Pluto {
	return Pluto{
		From:       strings.ToUpper(from),
		To:         ToPluto,
		isResponse: true,
		Response:   response,
	}
}

func NewPlutoResponseFrom(request Pluto, response string) Pluto {
	return Pluto{
		From:       request.From,
		To:         ToPluto,
		isResponse: true,
		Response:   response,
		Cmd:        request.Cmd,
	}
}

func (pl Pluto) Command() Command {
	return Command(pl)
}

func (pl Pluto) GetParams() []string {
	params := strings.Split(pl.Cmd, ":::")

	switch params[0] {
	case "GETURL":
		return params[1:]

	default:
		return params
	}
}

func (pl Pluto) GetURL(channelID string) Pluto {
	pl.Cmd = fmt.Sprintf("GETURL:::%s", channelID)
	return pl
}

func (pl Pluto) IsGetURL() bool {
	return strings.HasPrefix(pl.Cmd, "GETURL")
}
