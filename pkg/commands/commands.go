package commands

import "fmt"

const (
	ResponseOK   = "OK"
	ResponseFail = "FAIL"

	CommandStop = "STOP"

	ToMediaPlaylist  = "MEDIAPLAYLIST"
	ToMasterPlaylist = "MASTERPLAYLIST"
	ToPluto          = "PLUTO"
)

type Command struct {
	From       string
	To         string
	Cmd        string
	isResponse bool
	Response   string
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
