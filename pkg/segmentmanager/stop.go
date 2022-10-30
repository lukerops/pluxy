package segmentmanager

import "os"

func (sm *segmentManager) Stop() error {
	sm.mutex.Lock()

	return os.RemoveAll(sm.Dir)
}
