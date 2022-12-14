package m3u8

import (
	"fmt"
	"strings"
)

// #EXTINF
type Segment struct {
	URI        string
	Duration   float64
	Name       string
	CustomTags map[string]string
	Key        *Key
}

func (segment *Segment) String() string {
	var customTags []string
	for key, value := range segment.CustomTags {
		customTags = append(customTags, fmt.Sprintf("%s=\"%s\"", key, value))
	}

	params := []string{
		fmt.Sprintf("%f", segment.Duration),
	}

	if len(customTags) > 0 {
		params = append(params, " ", strings.Join(customTags, " "))
	}

	params = append(params, ",")

	if len(segment.Name) > 0 {
		params = append(params, " ", segment.Name)
	}

	var outputParams []string
	if segment.Key != nil {
		outputParams = append(outputParams, segment.Key.String())
	}

	outputParams = append(outputParams, fmt.Sprintf("#EXTINF:%s\n%s", strings.Join(params, ""), segment.URI))
	return strings.Join(outputParams, "\n")
}
