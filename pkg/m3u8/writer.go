package m3u8

import (
	"fmt"
	"io"
	"strings"
	"text/template"
)

func WriteMasterPlaylist(playlist *MasterPlaylist, writer io.Writer) error {
    tmpl := template.Must(
        template.New("MasterPlaylist").Parse(
            "#EXTM3U\n{{ range .Medias }}{{.}}{{end}}\n{{ range .Streams }}{{.}}{{end}}\n",
        ),
    )

    if err := tmpl.Execute(writer, playlist); err != nil {
        return err
    }

    return nil
}

func (stream *Stream) String() string {
    params := []string{
        fmt.Sprintf("PROGRAM-ID=%d", stream.ProgramID),
        fmt.Sprintf("BANDWIDTH=%d", stream.Bandwidth),
        fmt.Sprintf("SUBTITLES=\"%s\"", stream.Subtitles),
    }

    return fmt.Sprintf("#EXT-X-STREAM-INF:%s\n%s", strings.Join(params, ","), stream.URI)
}

func (media *Media) String() string {
    var isDefault, isForced string

    if media.Default {
        isDefault = "YES"
    } else {
        isDefault = "NO"
    }

    if media.Forced {
        isForced = "YES"
    } else {
        isForced = "NO"
    }

    params := []string{
        fmt.Sprintf("TYPE=%s", media.Type),
        fmt.Sprintf("GROUP-ID=\"%s\"", media.GroupID),
        fmt.Sprintf("NAME=\"%s\"", media.Name),
        fmt.Sprintf("DEFAULT=%s", isDefault),
        fmt.Sprintf("FORCED=%s", isForced),
        fmt.Sprintf("URI=\"%s\"", media.URI),
        fmt.Sprintf("LANGUAGE=\"%s\"", media.Language),
    }

    return fmt.Sprintf("#EXT-X-MEDIA:%s", strings.Join(params, ","))
}
