package m3u8

import (
	"fmt"
	"strings"
)

func bool2String(val bool) string {
    if val {
        return "YES"
    }
    return "NO"
}

// #EXT-X-MEDIA
type Media struct {
	Type     string
	GroupID  string
	Name     string
	Default  bool
	Forced   bool
	URI      string
	Language string
}

func (media *Media) String() string {
	params := []string{
		fmt.Sprintf("TYPE=%s", media.Type),
		fmt.Sprintf("GROUP-ID=\"%s\"", media.GroupID),
		fmt.Sprintf("NAME=\"%s\"", media.Name),
		fmt.Sprintf("DEFAULT=%s", bool2String(media.Default)),
		fmt.Sprintf("FORCED=%s", bool2String(media.Forced)),
		fmt.Sprintf("URI=\"%s\"", media.URI),
		fmt.Sprintf("LANGUAGE=\"%s\"", media.Language),
	}

	return fmt.Sprintf("#EXT-X-MEDIA:%s", strings.Join(params, ","))
}
