package segmentmanager

import (
	"fmt"
	"os"

	"github.com/gofrs/uuid"
)

func (sm *segmentManager) Add(channelID, segmentURI string, duration float64, data []byte) error {
	if sm.Check(channelID, segmentURI) {
		return nil
	}

	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	segments, ok := sm.index[channelID]
	if !ok {
		segments = make([]segmentInfo, 0)
	}

	id, err := uuid.NewV4()
	if err != nil {
		return err
	}

	filename := fmt.Sprintf("%s/%s_%s.ts", sm.dir, channelID, id)
	if err := os.WriteFile(filename, data, 0744); err != nil {
		return err
	}

	segments = append(segments, segmentInfo{
		uri:      segmentURI,
		filename: filename,
		duration: duration,
	})

	if len(segments) > sm.maxSegments {
		if err := os.Remove(segments[0].filename); err != nil {
			return err
		}
		segments = segments[1:]
	}

	sm.index[channelID] = segments
	return nil
}
