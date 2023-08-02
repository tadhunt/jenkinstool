package jenkinstool

import (
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/number"
	"gopkg.in/vansante/go-dl-stream.v2"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

const (
	Esc       = "\u001B["
	EraseLine = Esc + "2K"
	SOL       = "\r"
)

type BuildMetadata struct {
	ID        *string     `json:"id"`
	Result    *string     `json:"result"`
	Artifacts []*Artifact `json:"artifacts"`
}

type Artifact struct {
	DisplayPath  string `json:"displayPath"`
	Filename     string `json:"fileName"`
	RelativePath string `json:"relativePath"`
}

func GetBuild(src *url.URL, build string) (*BuildMetadata, error) {
	build = ParseBuild(build)
	metadata := &BuildMetadata{}

	err := GetBuildMetadata(src, build, metadata)
	if err != nil {
		return nil, err
	}

	return metadata, nil
}

func ParseBuild(build string) string {
	switch build {
	case "":
		fallthrough
	case "latest":
		return "lastSuccessfulBuild"
	default:
		return build
	}
}

func GetBuildMetadata(src *url.URL, build string, metadata any) error {
	u := fmt.Sprintf("%s/%s/api/json", src.String(), build)

	response, err := http.Get(u)
	if err != nil {
		return err
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	s, isString := metadata.(*string)
	if isString {
		*s = string(body)
	} else {
		err = json.Unmarshal(body, metadata)
		if err != nil {
			return err
		}
	}

	return nil
}

type StatusWriter struct {
	p      *message.Printer
	format number.FormatFunc
	last   int64
	total  int64
	start  time.Time
	name   string
	quiet  bool
}

func (sw *StatusWriter) Write(data []byte) (int, error) {
	sw.total += int64(len(data))

	if !sw.quiet {
		if sw.total-sw.last >= 256*1000 {
			kb := float64(sw.total) / 1000.0
			elapsed := time.Now().Sub(sw.start)
			kbps := kb / elapsed.Seconds()
			sw.p.Fprintf(os.Stdout, "%s%sDownloading %s %v KB (%v KB/s)", EraseLine, SOL, sw.name, sw.format(kb), sw.format(kbps))
			os.Stdout.Sync()
			sw.last = sw.total
		}
	}

	return len(data), nil
}

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
