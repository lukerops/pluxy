package downloader

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"time"

	"github.com/lukerops/pluxy/pkg/m3u8"
	"github.com/lukerops/pluxy/pkg/segmentmanager"
)

func (d *downloader) Run(ctx context.Context) {
	for {
		rawPlaylist, err := d.downloadFile(ctx, d.channelURL)
		if err != nil {
			fmt.Println("download master playlist failed; err:", err.Error())
			continue
		}

		playlist, err := m3u8.ReadMasterPlaylist(bytes.NewReader(rawPlaylist))
		if err != nil {
			fmt.Println("parse master playlist failed; err:", err.Error())
			continue
		}

		sort.Slice(playlist.Streams, func(i, j int) bool {
			return playlist.Streams[i].Bandwidth > playlist.Streams[j].Bandwidth
		})

		chURL, err := url.Parse(d.channelURL)
		if err != nil {
			panic(fmt.Sprintf("parse channel url failed; err: %s", err.Error()))
		}

		streamURL, err := chURL.Parse(playlist.Streams[0].URI)
		if err != nil {
			fmt.Println("parse stream url failed; err:", err.Error())
			continue
		}

		d.runMediaPlaylist(ctx, streamURL.String())
	}
}

func (d *downloader) runMediaPlaylist(ctx context.Context, streamURL string) {
	var lastSeqNo uint64 = 0

	for {
		rawPlaylist, err := d.downloadFile(ctx, streamURL)
		if err != nil {
			fmt.Println("download media playlist failed; err:", err.Error())
			break
		}

		playlist, err := m3u8.ReadMediaPlaylist(bytes.NewReader(rawPlaylist))
		if err != nil {
			fmt.Println("parse media playlist failed; err:", err.Error())
			break
		}

		if playlist.SeqNo == nil || (playlist.SeqNo != nil && *playlist.SeqNo == lastSeqNo) {
			continue
		}

		for _, segment := range playlist.Segments {
			if segmentmanager.SegmentManager.Check(d.channelID, segment.URI) {
				continue
			}

			pattern := regexp.MustCompile(`_ad/creative/|dai\.google\.com|Pluto_TV_OandO/.*Bumper`)
			if pattern.MatchString(segment.URI) {
				fmt.Println("Filtering Ads; url:", segment.URI)
				return
			}

			fmt.Println("Downloading segment:", segment.URI)

			segData, err := d.downloadSegment(ctx, segment)
			if err != nil {
				fmt.Println("download segment failed; err:", err.Error())
				continue
			}

			segmentmanager.SegmentManager.Add(d.channelID, segment.URI, segment.Duration, segData)
		}

		time.Sleep(time.Duration(playlist.Segments[0].Duration) * time.Second)
	}
}
