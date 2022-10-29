package downloader

import (
	"context"
	"io/ioutil"

	"github.com/lukerops/pluxy/pkg/m3u8"
)

func (d *downloader) DownloadSegment(ctx context.Context, segment *m3u8.Segment, filename string) error {
	segData, err := d.DownloadFile(ctx, segment.URI)
	if err != nil {
		return err
	}

	if segment.Key != nil {
		keyData, err := d.DownloadFile(ctx, segment.Key.URI)
		if err != nil {
			return err
		}

		// Remove o 0x
		iv := segment.Key.IV[2:]

		segData, err = decryptAesCBC(segData, keyData, []byte(iv))
		if err != nil {
			return err
		}
	}

	if err := ioutil.WriteFile(filename, segData, 0744); err != nil {
		return err
	}
	return nil
}
