package main

import(
	"fmt"
	"context"
	"regexp"
	"os"
	"io"
	"net/url"
	"net/http"
	"time"
	"encoding/json"
	"github.com/integrii/flaggy"
	"gopkg.in/vansante/go-dl-stream.v2"
	"golang.org/x/text/message"
	"golang.org/x/text/number"
	"golang.org/x/text/language"
)

const (
	Esc       = "\u001B["
	EraseLine = Esc + "2K"
	SOL       = "\r"
)

type Cmd struct {
	cmd     *flaggy.Subcommand
	handler func(cmd *Cmd) error
}

type BuildMetadata struct {
	ID        *string     `json:"id"`
	Artifacts []*Artifact `json:"artifacts"`
}

type Artifact struct {
      DisplayPath  string `json:"displayPath"`
      Filename     string `json:"fileName"`
      RelativePath string `json:"relativePath"`
}

type StatusWriter struct {
	p      *message.Printer
	format number.FormatFunc
	last   int64
	total  int64
	start  time.Time
	name   string
}

var (
	serverURL *url.URL
)

func main() {

	flaggy.SetName(os.Args[0])
	flaggy.SetDescription("Tool for interacting with the Jenkins API")
	flaggy.DefaultParser.AdditionalHelpPrepend = "https://github.com/tadhunt/jenkinstool"
	flaggy.SetVersion("0.1")

	server := ""
	flaggy.String(&server, "s", "server", "[required] URL of Jenkins server to interact with")

	cmds := []*Cmd{
		newGetCmd(),
		newDownloadCmd(),
	}

	for _, cmd := range cmds {
		flaggy.AttachSubcommand(cmd.cmd, 1)
	}

	flaggy.Parse()

	if (server == "") {
		flaggy.DefaultParser.ShowHelpWithMessage("-server is required")
		return
	}

	var err error
	serverURL, err = url.Parse(server)
	if err != nil {
		flaggy.DefaultParser.ShowHelpWithMessage(fmt.Sprintf("parse url: %v", err))
		return
	}


	for _, cmd := range cmds {
		if cmd.cmd.Used {
			err := cmd.handler(cmd)
			if err != nil {
				flaggy.DefaultParser.ShowHelpWithMessage(fmt.Sprintf("cmd %s: %v", cmd.cmd.Name, err))
			}
			return
			
		}
	}
}

func newGetCmd() *Cmd {
	build := ""

	get := flaggy.NewSubcommand("get")
	get.Description = "Get Build Info"

	get.String(&build, "b", "build", "[optional] Build to fetch (defaults to latest)")

	handler := func(cmd *Cmd) error {
		build = parseBuild(build)
		metadata, err := getBuildMetadata(build)
		if err != nil {
			return err
		}

		id := "<unknown>"
		if metadata.ID != nil {
			id = *metadata.ID
		}

		fmt.Printf("Build    %s\n", build)
		fmt.Printf("ID       %v\n", id)

		for _, artifact := range metadata.Artifacts {
			fmt.Printf("Artifact %s\n", artifact.DisplayPath)
		}

		return nil
	}

	return &Cmd{cmd: get, handler: handler}
}

func newDownloadCmd() *Cmd {
	build := ""
	artifactFilter := ""
	dstdir := ""
	replace := false

	get := flaggy.NewSubcommand("download")
	get.Description = "download build artifact"

	get.String(&build, "b", "build", "[optional] Build to fetch (defaults to latest)")
	get.String(&artifactFilter, "a", "artifact", "[optional] regex specifying which artifacts to fetch (default all)")
	get.String(&dstdir, "d", "dstdir", "[optional] Destination directory to download artifact(s) into")
	get.Bool(&replace, "r", "replace", "[optional] replace artifacts if they already exist")

	handler := func(cmd *Cmd) error {
		build = parseBuild(build)

		if artifactFilter == "" {
			artifactFilter = ".*"
		}

		artifactRe, err := regexp.Compile(artifactFilter)
		if err != nil {
			return err
		}

		if dstdir == "" {
			dstdir = "."
		}
		

		st, err := os.Stat(dstdir)
		if os.Stat(dstdir); err != nil {
			return fmt.Errorf("%s: %v", dstdir, err)
		}
		if !st.IsDir() {
			return fmt.Errorf("%s: is not a directory", dstdir)
		}

		metadata, err := getBuildMetadata(build)
		if err != nil {
			return err
		}

		for  _, artifact := range metadata.Artifacts {
			if artifactRe.MatchString(artifact.DisplayPath) {
				err = download(build, artifact, dstdir, replace)
				if err != nil {
					return err
				}
			}
		}

		return nil
	}

	return &Cmd{cmd: get, handler: handler}
}

func download(build string, artifact *Artifact, dstdir string, replace bool) error {
	src := fmt.Sprintf("%s/%s/artifact/%s", serverURL.String(), build, artifact.RelativePath)
	dst := fmt.Sprintf("%s/%s", dstdir, artifact.Filename)

	_, err := os.Stat(dst)
	if err == nil {
		if !replace {
			return fmt.Errorf("%s: already exists and -replace not specified")
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
		last: 0,
		total: 0,
		start: time.Now(),
		name: dst,
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

func getBuildMetadata(build string) (*BuildMetadata, error) {
	u := fmt.Sprintf("%s/%s/api/json", serverURL.String(), build)

	fmt.Printf("GET %s\n", u)

	response, err := http.Get(u)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	md := &BuildMetadata{}
	err = json.Unmarshal(body, &md)
	if err != nil {
		return nil, err
	}

	return md, nil
}

func parseBuild(build string) string {
	switch build {
	case "":
		fallthrough
	case "latest":
		return "lastSuccessfulBuild"
	default:
		return build
	}
}

func (sw *StatusWriter) Write(data []byte) (int, error) {
	sw.total += int64(len(data))

	if sw.total - sw.last >= 256*1000 {
		kb := float64(sw.total) / 1000.0
		elapsed := time.Now().Sub(sw.start)
		kbps := kb / elapsed.Seconds()
		sw.p.Fprintf(os.Stdout, "%s%sDownloading %s %v KB (%v KB/s)", EraseLine, SOL, sw.name, sw.format(kb), sw.format(kbps))
		os.Stdout.Sync()
		sw.last = sw.total
	}

	return len(data), nil
}
