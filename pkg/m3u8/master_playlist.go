package m3u8

import "strings"

type MasterPlaylist struct {
	Streams []*Stream // #EXT-X-STREAM-INF
	Medias  []*Media  // #EXT-X-MEDIA
}

func (playlist *MasterPlaylist) String() string {
	params := []string{"#EXTM3U"}

	for _, media := range playlist.Medias {
		params = append(params, media.String())
	}

	for _, stream := range playlist.Streams {
		params = append(params, stream.String())
	}

	return strings.Join(params, "\n")
}
