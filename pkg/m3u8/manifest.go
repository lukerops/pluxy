package m3u8

import (
	"fmt"
	"strings"
	"time"
)

type Manifest struct {
	// Master Manifest
	Streams []*Stream // #EXT-X-STREAM-INF
	Medias  []*Media  // #EXT-X-MEDIA

	// Media Manifest
	Segments         []*Segment // #EXTINF
	SeqNo            *uint64    // #EXT-X-MEDIA-SEQUENCE
	DiscontinuitySeq *uint64    // #EXT-X-DISCONTINUITY-SEQUENCE
	TargetDuration   *float64   // #EXT-X-TARGETDURATION
	ProgramDateTime  *time.Time // #EXT-X-PROGRAM-DATE-TIME
}

func (m *Manifest) IsMaster() bool {
	return len(m.Streams) > 0
}

func (m *Manifest) IsMedia() bool {
	return len(m.Segments) > 0
}

func (m *Manifest) String() string {
	params := []string{"#EXTM3U"}

	// Master Manifest
	if m.IsMaster() {
		for _, media := range m.Medias {
			params = append(params, media.String())
		}

		for _, stream := range m.Streams {
			params = append(params, stream.String())
		}

		return strings.Join(params, "\n")
	}

	// Media Manifest
	if m.SeqNo != nil {
		params = append(params, fmt.Sprintf("#EXT-X-MEDIA-SEQUENCE:%d", *m.SeqNo))
	}

	if m.DiscontinuitySeq != nil {
		params = append(params, fmt.Sprintf("#EXT-X-DISCONTINUITY-SEQUENCE:%d", *m.DiscontinuitySeq))
	}

	if m.TargetDuration != nil {
		params = append(params, fmt.Sprintf("#EXT-X-TARGETDURATION:%f", *m.TargetDuration))
	}

	if m.ProgramDateTime != nil {
		params = append(params, fmt.Sprintf(
			"#EXT-X-PROGRAM-DATE-TIME:%s",
			m.ProgramDateTime.Format("2006-01-02T15:04:05.000Z"),
		))
	}

	for _, segment := range m.Segments {
		params = append(params, segment.String())
	}

	return strings.Join(params, "\n")
}
