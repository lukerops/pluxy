package m3u8

import (
	"time"
)

type MasterPlaylist struct {
	Streams []*Stream // #EXT-X-STREAM-INF
    Medias []*Media // #EXT-X-MEDIA
}

// #EXT-X-STREAM-INF
type Stream struct {
	URI          string
	ProgramID    int32
	Bandwidth    int32
	Subtitles    string
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

type MediaPlaylist struct {
	Segments         []*Segment // #EXTINF
	SeqNo            uint64     // #EXT-X-MEDIA-SEQUENCE
	DiscontinuitySeq uint64     // #EXT-X-DISCONTINUITY-SEQUENCE
	TargetDuration   float64    // #EXT-X-TARGETDURATION
	ProgramDateTime  *time.Time // #EXT-X-PROGRAM-DATE-TIME
}

// #EXTINF
type Segment struct {
	URI        string
	Duration   float64
	Name       string
	CustomTags map[string]string
	Key        *Key
}

// #EXT-X-KEY
type Key struct {
	Method string
	URI    string
	IV     string
}
