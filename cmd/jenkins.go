package main

import (
	"fmt"
	"github.com/integrii/flaggy"
	"github.com/tadhunt/jenkinstool"
	"net/url"
	"os"
	"regexp"
)

type Cmd struct {
	cmd     *flaggy.Subcommand
	handler func(cmd *Cmd) error
}

var (
	serverURL *url.URL
	quiet     = false
)

func main() {
	flaggy.SetName(os.Args[0])
	flaggy.SetDescription("Tool for interacting with the Jenkins API")
	flaggy.DefaultParser.AdditionalHelpPrepend = "https://github.com/tadhunt/jenkinstool"
	flaggy.SetVersion("0.1")

	server := ""
	flaggy.String(&server, "s", "server", "[required] URL of Jenkins server to interact with")
	flaggy.Bool(&quiet, "q", "quiet", "[optional] don't print extra info")

	cmds := []*Cmd{
		newGetCmd(),
		newDownloadCmd(),
	}

	for _, cmd := range cmds {
		flaggy.AttachSubcommand(cmd.cmd, 1)
	}

	flaggy.Parse()

	if server == "" {
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
	rawJson := false

	get := flaggy.NewSubcommand("get")
	get.Description = "Get Build Metadata"

	get.String(&build, "b", "build", "[optional] Build to fetch (defaults to latest)")
	get.Bool(&rawJson, "j", "json", "[optional] dump all of the json metadata")

	handler := func(cmd *Cmd) error {
		build = jenkinstool.ParseBuild(build)
		if rawJson {
			metadata, err := jenkinstool.GetRawBuildMetadata(serverURL, build)
			if err != nil {
				return err
			}
			fmt.Printf("%s\n", metadata)
		} else {
			metadata, err := jenkinstool.GetBuildMetadata(serverURL, build)
			if err != nil {
				return err
			}

			fmt.Printf("Build    %s\n", build)
			fmt.Printf("ID       %v\n", jenkinstool.String(metadata.ID))
			fmt.Printf("Result   %v\n", jenkinstool.String(metadata.Result))

			for _, artifact := range metadata.Artifacts {
				fmt.Printf("Artifact %s\n", artifact.DisplayPath)
			}
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
		build = jenkinstool.ParseBuild(build)

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

		metadata, err := jenkinstool.GetBuildMetadata(serverURL, build)
		if err != nil {
			return err
		}

		for _, artifact := range metadata.Artifacts {
			if artifactRe.MatchString(artifact.DisplayPath) {
				err = jenkinstool.Download(serverURL, build, artifact, dstdir, replace, quiet)
				if err != nil {
					return err
				}
			}
		}

		return nil
	}

	return &Cmd{cmd: get, handler: handler}
}
