package channel

import (
	"fmt"
	"path"

	"github.com/lukerops/pluxy/pkg/m3u8"
	"github.com/lukerops/pluxy/pkg/segmentmanager"
)

func (ch *channelService) GetChannelPlaylist(channelID string) (string, error) {
	filenames, durations, err := segmentmanager.SegmentManager.Get(channelID)
	if err != nil {
		return "", err
	}
	if len(filenames) == 0 {
		return "", ErrChannelNotFound
	}

	var segmentStart, segmentEnd uint

	switch size := len(filenames); {
	case size < 5:
		segmentStart = 0
		segmentEnd = uint(size)
	case size >= 5 && size < 10:
		segmentStart = uint(size) - 5
		segmentEnd = uint(size)
	case size >= 10:
		segmentStart = 4
		segmentEnd = 10
	}

	filenames = filenames[segmentStart:segmentEnd]
	durations = durations[segmentStart:segmentEnd]

	if filenames[0] != ch.lastFirstSegmentFilename {
		ch.lastSeqNo += 1
		ch.lastFirstSegmentFilename = filenames[0]
	}

	segments := make([]*m3u8.Segment, len(filenames))
	for index := range filenames {
		segments[index] = &m3u8.Segment{
			URI:      fmt.Sprintf("/segments/%s", path.Base(filenames[index])),
			Duration: durations[index],
		}
	}

	playlist := m3u8.MediaPlaylist{
		SeqNo:    &ch.lastSeqNo,
		Segments: segments,
	}

	return playlist.String(), nil
}
