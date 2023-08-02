package jenkinstool

import (
	"context"
	"fmt"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/number"
	"gopkg.in/vansante/go-dl-stream.v2"
	"net/url"
	"os"
	"time"
)

func Download(serverURL *url.URL, build string, artifact *Artifact, dstdir string, replace bool, quiet bool) error {
	src := fmt.Sprintf("%s/%s/artifact/%s", serverURL.String(), build, artifact.RelativePath)
	dst := fmt.Sprintf("%s/%s", dstdir, artifact.Filename)

	_, err := os.Stat(dst)
	if err == nil {
		if !replace {
			return fmt.Errorf("%s: already exists and -replace not specified", dst)
		}

		err = os.Remove(dst)
		if err != nil {
			return fmt.Errorf("remove %s: %v", dst, err)
		}
	} else {
		if !os.IsNotExist(err) {
			return fmt.Errorf("stat %s: %v", dst, err)
		}
	}

	sw := &StatusWriter{
		p:      message.NewPrinter(language.English),
		format: number.NewFormat(number.Decimal, number.MaxFractionDigits(2), number.MinFractionDigits(2)),
		last:   0,
		total:  0,
		start:  time.Now(),
		name:   dst,
		quiet:  quiet,
	}

	err = dlstream.DownloadStream(context.Background(), src, dst, sw)
	if err != nil {
		return err
	}

	elapsed := time.Now().Sub(sw.start)
	kbps := float64(sw.total) / 1000.0 / elapsed.Seconds()

	sw.p.Printf("%s%sDownloaded %s %v bytes (%v KB/s)\n", EraseLine, SOL, dst, number.Decimal(sw.total), sw.format(kbps))

	return nil
}
