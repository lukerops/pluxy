package m3u8

import (
	"fmt"
	"strings"
)

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
	var isDefault, isForced string

	if media.Default {
		isDefault = "YES"
	} else {
		isDefault = "NO"
	}

	if media.Forced {
		isForced = "YES"
	} else {
		isForced = "NO"
	}

	params := []string{
		fmt.Sprintf("TYPE=%s", media.Type),
		fmt.Sprintf("GROUP-ID=\"%s\"", media.GroupID),
		fmt.Sprintf("NAME=\"%s\"", media.Name),
		fmt.Sprintf("DEFAULT=%s", isDefault),
		fmt.Sprintf("FORCED=%s", isForced),
		fmt.Sprintf("URI=\"%s\"", media.URI),
		fmt.Sprintf("LANGUAGE=\"%s\"", media.Language),
	}

	return fmt.Sprintf("#EXT-X-MEDIA:%s", strings.Join(params, ","))
}
