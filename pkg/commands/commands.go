package commands

import (
	"fmt"
	"strings"
)

type CommandHandler string

func (h CommandHandler) IsProvider() bool {
	return strings.HasPrefix(string(h), "PROVIDER")
}

const (
	ResponseOK   = "OK"
	ResponseFail = "FAIL"

	CommandStop = "STOP"

	HlsDownloader CommandHandler = "HLSDOWNLOADER"
	PlutoProvider CommandHandler = "PROVIDER:PLUTO"
)

type Command struct {
	From       CommandHandler
	To         CommandHandler
	Cmd        string
	isResponse bool
	Response   string
}

func NewResponseFrom(cmd Command, response string) Command {
	cmd.isResponse = true
	cmd.Response = response
	return cmd
}

func (cmd Command) String() string {
	if cmd.isResponse {
		return fmt.Sprintf("%s:::%s:::RESPONSE:::%s:::%s", cmd.From, cmd.To, cmd.Cmd, cmd.Response)
	}
	return fmt.Sprintf("%s:::%s:::REQUEST:::%s", cmd.From, cmd.To, cmd.Cmd)
}

func (cmd Command) IsRequest() bool {
	return !cmd.isResponse
}

func (cmd Command) IsResponse() bool {
	return cmd.isResponse
}
