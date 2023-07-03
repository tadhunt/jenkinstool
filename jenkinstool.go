package main

import(
	"fmt"
	"context"
	"regexp"
	"os"
	"io"
	"net/url"
	"net/http"
	"encoding/json"
	"github.com/integrii/flaggy"
	"gopkg.in/vansante/go-dl-stream.v2"
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
	total int64
}

var (
	serverURL *url.URL
)

func main() {

	flaggy.SetName("Jenkins CLI Tool")
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
				return
			}
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

	get := flaggy.NewSubcommand("download")
	get.Description = "download build artifact"

	get.String(&build, "b", "build", "[optional] Build to fetch (defaults to latest)")
	get.String(&artifactFilter, "a", "artifact", "[optional] regex specifying which artifacts to fetch (default all)")
	get.String(&dstdir, "d", "dstdir", "[optional] Destination directory to download artifact(s) into")

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
				err = download(build, artifact, dstdir)
				if err != nil {
					return err
				}
			}
		}

		return nil
	}

	return &Cmd{cmd: get, handler: handler}
}

func download(build string, artifact *Artifact, dstdir string) error {
	src := fmt.Sprintf("%s/%s/artifact/%s", serverURL.String(), build, artifact.RelativePath)
	dst := fmt.Sprintf("%s/%s", dstdir, artifact.Filename)

	fmt.Printf("downloading %s to %s\n", src, dst)
	sw := &StatusWriter{}

	err := dlstream.DownloadStream(context.Background(), src, dst, sw)
	if err != nil {
		return err
	}

	fmt.Printf("\nDownloaded %s (%d bytes)\n", artifact.Filename, sw.total)

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
	fmt.Fprint(os.Stdout, "%s%s%d", EraseLine, SOL, sw.total)
	os.Stdout.Sync()

	return len(data), nil
}
