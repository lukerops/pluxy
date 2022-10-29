package m3u8

import (
	"fmt"
	"strings"
	"time"
)

type MediaPlaylist struct {
	Segments         []*Segment // #EXTINF
	SeqNo            *uint64    // #EXT-X-MEDIA-SEQUENCE
	DiscontinuitySeq *uint64    // #EXT-X-DISCONTINUITY-SEQUENCE
	TargetDuration   *float64   // #EXT-X-TARGETDURATION
	ProgramDateTime  *time.Time // #EXT-X-PROGRAM-DATE-TIME
}

func (playlist *MediaPlaylist) String() string {
    params := []string{"#EXTM3U"}

    if playlist.SeqNo != nil {
        params = append(params, fmt.Sprintf("#EXT-X-MEDIA-SEQUENCE:%d", *playlist.SeqNo))
    }

    if playlist.DiscontinuitySeq != nil {
        params = append(params, fmt.Sprintf("#EXT-X-DISCONTINUITY-SEQUENCE:%d", *playlist.DiscontinuitySeq))
    }

    if playlist.TargetDuration != nil {
        params = append(params, fmt.Sprintf("#EXT-X-TARGETDURATION:%f", *playlist.TargetDuration))
    }

    if playlist.ProgramDateTime != nil {
        params = append(params, fmt.Sprintf(
            "#EXT-X-PROGRAM-DATE-TIME:%s",
            playlist.ProgramDateTime.Format("2006-01-02T15:04:05.000Z"),
        ))
    }

    for _, segment := range playlist.Segments {
        params = append(params, segment.String())
    }

    return strings.Join(params, "\n")
}
