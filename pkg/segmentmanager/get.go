package segmentmanager

func (sm *segmentManager) Get(channelID string) []string {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	segments := sm.index[channelID]

	filenames := make([]string, len(segments))
	for index, segment := range segments {
		filenames[index] = segment.filename
	}

	return filenames
}
