package downloader

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/lukerops/pluxy/pkg/m3u8"
	"github.com/lukerops/pluxy/pkg/segmentmanager"
	"github.com/rs/zerolog/log"
)

func (d *downloader) mediaPlaylistWorker(workerChTx chan<- string, workerChRx <-chan string) {
	logger := log.With().Str("module", "downloader").Str("worker", "mediaPlaylist").Logger()

	channels := make(map[string]string)
	internalCh := make(chan string, 50)

	var segments []*m3u8.Segment

	var lastSeqNo uint64 = 0

	for {
		select {
		case <-d.mediaPlaylistStopCh:
			return

		case msg := <-workerChRx:
			logger.Info().Str("msg", msg).Msg("chegou uma mensagem")
			params := strings.Split(msg, ":::")

			switch {
			// REGISTER:::{CHANNEL ID}
			case strings.HasPrefix(msg, "REGISTER"):

				if len(params) != 2 {
					workerChTx <- "FAILED"
				}

				channels[params[1]] = ""
				workerChTx <- fmt.Sprintf("RESPONSE:::%s:::OK", msg)
				workerChTx <- fmt.Sprintf("GETURL:::%s", params[1])

			// RESPONSE:::GETURL:::{CHANNEL ID}:::{RESPONSE}
			case strings.HasPrefix(msg, "RESPONSE:::GETURL"):
				if len(params) != 4 {
					logger.Error().Str("command", msg).Msg("invalid command")
				}

				if _, ok := channels[params[2]]; !ok {
					logger.Error().Str("command", msg).Msg("invalid channel")
				}

				if params[3] == "FAILED" {
					workerChTx <- fmt.Sprintf("GETURL:::%s", params[2])
					continue
				}

				channels[params[2]] = params[3]
				internalCh <- fmt.Sprintf("DOWNLOAD:::PLAYLIST:::%s:::%s", params[2], params[3])
			}

		case msg := <-internalCh:
			logger.Info().Str("msg", msg).Msg("chegou uma mensagem")
			params := strings.Split(msg, ":::")

			switch {
			// DOWNLOAD:::PLAYLIST:::{CHANNEL ID}:::{PLAYLIST URL}
			case strings.HasPrefix(msg, "DOWNLOAD:::PLAYLIST"):
				if len(params) != 4 {
					logger.Error().Str("command", msg).Msg("invalid command")
				}

				ctx := context.Background()

				rawPlaylist, err := d.downloadFile(ctx, params[3])
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
					internalCh <- msg
					continue
				}

				for _, segment := range playlist.Segments {
					// key := ":::"
					// if segment.Key != nil {
					//     key = fmt.Sprintf("%s:::%s", segment.Key.URI, segment.Key.IV)
					// }

					segments = append(segments, segment)
					internalCh <- fmt.Sprintf(
						//"DOWNLOAD:::SEGMENT:::%s:::%s:::%f:::%s", params[2], segment.URI, segment.Duration, key)
						"DOWNLOAD:::SEGMENT:::%s:::%d", params[2], len(segments)-1)
				}

				internalCh <- msg

			// DOWNLOAD:::SEGMENT:::{CHANNEL ID}:::{SEGMENT URL}:::{SEGMENT DURATION}:::{KEY URL}:::{KEY IV}
			// DOWNLOAD:::SEGMENT:::{CHANNEL ID}:::{SEGMENT INDEX}
			case strings.HasPrefix(msg, "DOWNLOAD:::SEGMENT"):
				if len(params) != 4 {
					logger.Error().Str("command", msg).Msg("invalid command")
				}

				segmentIndex, _ := strconv.ParseInt(params[3], 10, 32)
				segment := segments[segmentIndex]

				if segmentmanager.SegmentManager.Check(params[2], segment.URI) {
					continue
				}

				pattern := regexp.MustCompile(`_ad/creative/|dai\.google\.com|Pluto_TV_OandO/.*Bumper`)
				if pattern.MatchString(segment.URI) {
					fmt.Println("Filtering Ads; url:", params)
					return
				}

				fmt.Println("Downloading segment:", segment.URI)

				ctx := context.Background()
				segData, err := d.downloadSegment(ctx, segment)
				if err != nil {
					fmt.Println("download segment failed; err:", err.Error())
					continue
				}

				segmentmanager.SegmentManager.Add(params[2], segment.URI, segment.Duration, segData)
			}
		}
	}
}
