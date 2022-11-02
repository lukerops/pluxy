package commands

import (
	"fmt"
	"strings"
)

type ProviderCmd struct {
	Command
}

func NewProviderCmd(from, to CommandHandler) ProviderCmd {
	return ProviderCmd{
		Command{
			From: from,
			To:   to,
		},
	}
}

func NewProviderCmdFrom(cmd Command) ProviderCmd {
	return ProviderCmd{cmd}
}

func NewProviderCmdResponse(from, to CommandHandler, response string) ProviderCmd {
	return ProviderCmd{
		Command{
			From:       from,
			To:         to,
			isResponse: true,
			Response:   response,
		},
	}
}

func NewProviderCmdResponseFrom(cmd ProviderCmd, response string) ProviderCmd {
	cmd.isResponse = true
	cmd.Response = response
	return cmd
}

func (cmd ProviderCmd) GetCommand() Command {
	return cmd.Command
}

func (cmd ProviderCmd) GetURL(channelID string) ProviderCmd {
	cmd.Cmd = fmt.Sprintf("GETURL:::%s", channelID)
	return cmd
}

func (cmd ProviderCmd) GetURLParams() (channelID string) {
	return strings.Split(cmd.Cmd, ":::")[1]
}

func (cmd ProviderCmd) IsGetURL() bool {
	return strings.HasPrefix(cmd.Cmd, "GETURL")
}
