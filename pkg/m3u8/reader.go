package m3u8

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func getLineParams(line string) map[string]string {
	params := make(map[string]string)

	pattern := regexp.MustCompile(`([a-zA-Z0-9_-]+)=("([^"]+)"|([^",]+))`)
	groups := pattern.FindAllStringSubmatch(line, -1)
	for _, group := range groups {
		key, value := group[1], group[3]
		if len(group[4]) > 0 {
			value = group[4]
		}

		params[key] = value
	}

	return params
}

type state struct {
	hasM3U bool
	stream *Stream

	key     *Key
	segment *Segment
}

func ReadManifest(text string) (*Manifest, error) {
    isMaster, err := regexp.MatchString(`(MEDIA:|STREAM-INF:)`, text)
    if err != nil {
        return nil, err
    }

    manifestReader := readMasterManifest
    if !isMaster {
        manifestReader = readMediaManifest
    }

    var buf bytes.Buffer
	_, err = buf.ReadFrom(strings.NewReader(text))
	if err != nil {
		return nil, err
	}

	state := new(state)
	manifest := new(Manifest)

	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			break
		}

		line = strings.Replace(line, "\n", "", 1)
		manifestReader(line, state, manifest)
    }

	if !state.hasM3U {
		return nil, errors.New("invalid M3U file")
	}

	return manifest, nil
}

func readMasterManifest(line string, state *state, manifest *Manifest) {
	switch {
	case strings.HasPrefix(line, "#EXTM3U"):
		state.hasM3U = true

	case strings.HasPrefix(line, "#EXT-X-MEDIA:"):
		media := new(Media)

		params := getLineParams(line)
		for key, value := range params {
			switch key {
			case "GROUP-ID":
				media.GroupID = value
			case "TYPE":
				media.Type = value
			case "NAME":
				media.Name = value
			case "DEFAULT":
				if value == "YES" {
					media.Default = true
				} else if value != "NO" {
					fmt.Println("Invalid #EXT-X-MEDIA param; key:", key, "; value:", value)
				}
			case "FORCED":
				if value == "YES" {
					media.Default = true
				} else if value != "NO" {
					fmt.Println("Invalid #EXT-X-MEDIA param; key:", key, "; value:", value)
				}
			case "URI":
				media.URI = value
			case "LANGUAGE":
				media.Language = value
			default:
				fmt.Println("Invalid #EXT-X-MEDIA param; key:", key)
			}
		}

		manifest.Medias = append(manifest.Medias, media)

	case strings.HasPrefix(line, "#EXT-X-STREAM-INF:"):
		state.stream = new(Stream)

		params := getLineParams(line)
		for key, value := range params {
			switch key {
			case "PROGRAM-ID":
				programID, _ := strconv.ParseInt(value, 10, 32)
				state.stream.ProgramID = int32(programID)
			case "BANDWIDTH":
				bandwidth, _ := strconv.ParseInt(value, 10, 32)
				state.stream.Bandwidth = int32(bandwidth)
			case "SUBTITLES":
				state.stream.Subtitles = value
			default:
				fmt.Println("Invalid #EXT-X-STREAM-INF param; key:", key)
			}
		}

	case !strings.HasPrefix(line, "#"):
		pattern := regexp.MustCompile(`(\S*)`)
		groups := pattern.FindStringSubmatch(line)

		if len(groups[0]) == 0 {
			return
		}

		if state.stream == nil {
			fmt.Println("Invalid URL Position")
		}

		state.stream.URI = line
		manifest.Streams = append(manifest.Streams, state.stream)
		state.stream = nil
	}
}

func readMediaManifest(line string, state *state, manifest *Manifest) {
	switch {
	case strings.HasPrefix(line, "#EXTM3U"):
		state.hasM3U = true

	case strings.HasPrefix(line, "#EXT-X-DISCONTINUITY-SEQUENCE:"):
		manifest.DiscontinuitySeq = new(uint64)
		fmt.Sscanf(line, "#EXT-X-DISCONTINUITY-SEQUENCE:%d", manifest.DiscontinuitySeq)

	case strings.HasPrefix(line, "#EXT-X-MEDIA-SEQUENCE:"):
		manifest.SeqNo = new(uint64)
		fmt.Sscanf(line, "#EXT-X-MEDIA-SEQUENCE:%d", manifest.SeqNo)

	case strings.HasPrefix(line, "#EXT-X-TARGETDURATION:"):
		manifest.TargetDuration = new(float64)
		fmt.Sscanf(line, "#EXT-X-TARGETDURATION:%f", manifest.TargetDuration)

	case strings.HasPrefix(line, "#EXT-X-PROGRAM-DATE-TIME:"):
		dateTime, _ := time.Parse("2006-01-02T15:04:05.000Z", line[25:])
		manifest.ProgramDateTime = &dateTime

	case strings.HasPrefix(line, "#EXT-X-KEY:"):
		state.key = new(Key)

		params := getLineParams(line)
		for key, value := range params {
			switch key {
			case "METHOD":
				state.key.Method = value
			case "URI":
				state.key.URI = value
			case "IV":
				state.key.IV = value
			}
		}

	case strings.HasPrefix(line, "#EXTINF:"):
		pattern := regexp.MustCompile(`^#EXTINF:(-?\d(\.\d+)?)([\S ]*), *([\S ]*)\n?$`)
		subStr := pattern.FindStringSubmatch(line)

		duration, _ := strconv.ParseFloat(subStr[1], 64)
		state.segment = new(Segment)
		state.segment.Duration = duration
		state.segment.Name = subStr[4]
		state.segment.CustomTags = getLineParams(subStr[3])
		state.segment.Key = state.key

	case !strings.HasPrefix(line, "#"):
		pattern := regexp.MustCompile(`(\S*)`)
		groups := pattern.FindStringSubmatch(line)

		if len(groups[0]) == 0 {
			return
		}

		if state.segment == nil {
			fmt.Println("Invalid URL Position")
		}

		state.segment.URI = line
		manifest.Segments = append(manifest.Segments, state.segment)

		state.key = nil
		state.segment = nil
	}
}
