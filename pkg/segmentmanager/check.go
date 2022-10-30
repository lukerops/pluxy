package segmentmanager

func (sm *segmentManager) Check(channelID, segmentURI string) bool {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	segments, ok := sm.index[channelID]
	if !ok {
		segments = make([]segmentInfo, 0)
	}

	for _, segInfo := range segments {
		if segInfo.uri == segmentURI {
			return true
		}
	}

	return false
}
