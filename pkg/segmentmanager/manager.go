package segmentmanager

import (
	"os"
	"strconv"
	"sync"
)

type segmentInfo struct {
	filename string
	uri      string
	duration float64
}

type segmentManager struct {
	maxSegments int
	index       map[string][]segmentInfo
	mutex       sync.RWMutex
	dir         string
}

var SegmentManager *segmentManager

func init() {
	maxSegments, err := strconv.ParseUint(os.Getenv("PLUXY_MAX_SEGMENTS"), 10, 32)
	if err != nil {
		maxSegments = 10
	}

	tmpDir, err := os.MkdirTemp("", "pluxy_")
	if err != nil {
		panic(err)
	}

	SegmentManager = &segmentManager{
		maxSegments: int(maxSegments),
		dir:         tmpDir,
		index:       make(map[string][]segmentInfo, 0),
	}
}
