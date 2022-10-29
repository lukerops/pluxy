package m3u8

import (
	"bytes"
	"errors"
	"fmt"
	"io"
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

type readMasterState struct {
	HasM3U       bool
	StreamMedias []*Media
	StreamOption *Stream
}

func ReadMasterPlaylist(reader io.Reader) (*MasterPlaylist, error) {
	buf := new(bytes.Buffer)

	_, err := buf.ReadFrom(reader)
	if err != nil {
		return nil, err
	}

	var (
		playlist MasterPlaylist
		state    readMasterState
	)
	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			break
		}

		line = strings.Replace(line, "\n", "", 1)
		switch {
		case strings.HasPrefix(line, "#EXTM3U"):
			state.HasM3U = true

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

			state.StreamMedias = append(state.StreamMedias, media)

		case strings.HasPrefix(line, "#EXT-X-STREAM-INF:"):
			state.StreamOption = new(Stream)

			params := getLineParams(line)
			for key, value := range params {
				switch key {
				case "PROGRAM-ID":
					programID, _ := strconv.ParseInt(value, 10, 32)
					state.StreamOption.ProgramID = int32(programID)
				case "BANDWIDTH":
					bandwidth, _ := strconv.ParseInt(value, 10, 32)
					state.StreamOption.Bandwidth = int32(bandwidth)
				case "SUBTITLES":
					state.StreamOption.Subtitles = value
					for _, media := range state.StreamMedias {
						if media.GroupID == state.StreamOption.Subtitles {
							state.StreamOption.StreamMedias = append(state.StreamOption.StreamMedias, media)
						}
					}

				default:
					fmt.Println("Invalid #EXT-X-STREAM-INF param; key:", key)
				}
			}

		case !strings.HasPrefix(line, "#"):
			pattern := regexp.MustCompile(`(\S*)`)
			groups := pattern.FindStringSubmatch(line)

			if len(groups[0]) == 0 {
				continue
			}

			if state.StreamOption == nil {
				fmt.Println("Invalid URL Position")
			}

			state.StreamOption.URI = line
			playlist.StreamOptions = append(playlist.StreamOptions, state.StreamOption)
			state.StreamOption = nil
		}
	}

	if !state.HasM3U {
		return nil, errors.New("invalid M3U file")
	}

	return &playlist, nil
}

type readMediaState struct {
	HasM3U  bool
	Key     *Key
	Segment *Segment
}

func ReadMediaPlaylist(reader io.Reader) (*MediaPlaylist, error) {
	buf := new(bytes.Buffer)

	_, err := buf.ReadFrom(reader)
	if err != nil {
		return nil, err
	}

	var (
		playlist MediaPlaylist
		state    readMediaState
	)
	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			break
		}

		line = strings.Replace(line, "\n", "", 1)
		switch {
		case strings.HasPrefix(line, "#EXTM3U"):
			state.HasM3U = true

		case strings.HasPrefix(line, "#EXT-X-DISCONTINUITY-SEQUENCE:"):
			_, err := fmt.Sscanf(line, "#EXT-X-DISCONTINUITY-SEQUENCE:%d", &playlist.DiscontinuitySeq)
			if err != nil {
				return nil, err
			}

		case strings.HasPrefix(line, "#EXT-X-MEDIA-SEQUENCE:"):
			_, err := fmt.Sscanf(line, "#EXT-X-MEDIA-SEQUENCE:%d", &playlist.SeqNo)
			if err != nil {
				return nil, err
			}

		case strings.HasPrefix(line, "#EXT-X-TARGETDURATION:"):
			_, err := fmt.Sscanf(line, "#EXT-X-TARGETDURATION:%f", &playlist.TargetDuration)
			if err != nil {
				return nil, err
			}

		case strings.HasPrefix(line, "#EXT-X-PROGRAM-DATE-TIME:"):
			dateTime, err := time.Parse("2006-01-02T15:04:05.000Z", line[25:])
			if err != nil {
				return nil, err
			}

			playlist.ProgramDateTime = &dateTime

		case strings.HasPrefix(line, "#EXT-X-KEY:"):
			state.Key = new(Key)

			params := getLineParams(line)
			for key, value := range params {
				switch key {
				case "METHOD":
					state.Key.Method = value
				case "URI":
					state.Key.URI = value
				case "IV":
					state.Key.IV = value
				}
			}

		case strings.HasPrefix(line, "#EXTINF:"):
			pattern := regexp.MustCompile(`^#EXTINF:(-?\d(\.\d+)?)([\S ]*), *([\S ]*)\n?$`)
			subStr := pattern.FindStringSubmatch(line)

			duration, err := strconv.ParseFloat(subStr[1], 64)
			if err != nil {
				return nil, err
			}

			state.Segment = new(Segment)
			state.Segment.Duration = duration
			state.Segment.Name = subStr[4]
			state.Segment.CustomTags = getLineParams(subStr[3])
			state.Segment.Key = state.Key

		case !strings.HasPrefix(line, "#"):
			pattern := regexp.MustCompile(`(\S*)`)
			groups := pattern.FindStringSubmatch(line)

			if len(groups[0]) == 0 {
				continue
			}

			if state.Segment == nil {
				fmt.Println("Invalid URL Position")
			}

			state.Segment.URI = line
			playlist.Segments = append(playlist.Segments, state.Segment)

			state.Key = nil
			state.Segment = nil
		}
	}

	if !state.HasM3U {
		return nil, errors.New("invalid M3U file")
	}

	return &playlist, nil
}
