package m3u8

import (
	"fmt"
	"strings"
)

// #EXT-X-STREAM-INF
type Stream struct {
	URI       string
	ProgramID int32
	Bandwidth int32
	Subtitles string
}

func (stream *Stream) String() string {
	params := []string{
		fmt.Sprintf("PROGRAM-ID=%d", stream.ProgramID),
		fmt.Sprintf("BANDWIDTH=%d", stream.Bandwidth),
		fmt.Sprintf("SUBTITLES=\"%s\"", stream.Subtitles),
	}

	return fmt.Sprintf("#EXT-X-STREAM-INF:%s\n%s", strings.Join(params, ","), stream.URI)
}
