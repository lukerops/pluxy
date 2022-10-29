package m3u8

import (
	"fmt"
	"strings"
)

// #EXT-X-KEY
type Key struct {
	Method string
	URI    string
	IV     string
}

func (key *Key) String() string {
    params := []string{
        fmt.Sprintf("METHOD=%s", key.Method),
        fmt.Sprintf("URI=\"%s\"", key.URI),
        fmt.Sprintf("IV=%s", key.IV),
    }

    return fmt.Sprintf("#EXT-X-KEY:%s", strings.Join(params, ","))
}
