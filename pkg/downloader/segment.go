package downloader

import (
	"context"

	"github.com/lukerops/pluxy/pkg/m3u8"
)

func (d *downloader) downloadSegment(ctx context.Context, segment *m3u8.Segment) ([]byte, error) {
	segData, err := d.downloadFile(ctx, segment.URI)
	if err != nil {
		return nil, err
	}

	if segment.Key != nil {
		keyData, err := d.downloadFile(ctx, segment.Key.URI)
		if err != nil {
			return nil, err
		}

		// Remove o 0x
		iv := segment.Key.IV[2:]

		segData, err = decryptAesCBC(segData, keyData, []byte(iv))
		if err != nil {
			return nil, err
		}
	}

	return segData, nil
}
