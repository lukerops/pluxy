package segmentmanager

import (
	"errors"
)

var ErrChannelNotFound = errors.New("channel not found")

func (sm *segmentManager) Get(channelID string) ([]string, []float64, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	segments, ok := sm.index[channelID]
	if !ok {
		return nil, nil, ErrChannelNotFound
	}

	filenames := make([]string, len(segments))
	durations := make([]float64, len(segments))
	for index, segment := range segments {
		filenames[index] = segment.filename
		durations[index] = segment.duration
	}

	return filenames, durations, nil
}
