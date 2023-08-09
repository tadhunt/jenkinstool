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
	since := ""
	showChanges := false
	rawJson := false

	get := flaggy.NewSubcommand("get")
	get.Description = "Get Build Metadata"

	get.String(&build, "", "build", "[optional] Build to fetch (defaults to latest)")
	get.Bool(&showChanges, "", "changes", "[optional] show changes")
	get.String(&build, "", "since", "[optional] Fetch all builds between the latest and this one")
	get.Bool(&rawJson, "", "json", "[optional] dump the raw json metadata")

	handler := func(cmd *Cmd) error {
		first :=  true
		var metadata *jenkinstool.BuildMetadata
		for {
			if !first {
				fmt.Printf("----------\n")
			}
			if rawJson {
				raw, err := jenkinstool.GetRawBuildMetadata(serverURL, build)
				if err != nil {
					return err
				}

				fmt.Printf("%s\n", raw)

				metadata, err = jenkinstool.GetBuildMetadataFromBytes(raw)
				if err != nil {
					return err
				}
			} else {
				var err error
				metadata, err = jenkinstool.GetBuildMetadata(serverURL, build)
				if err != nil {
					return err
				}

				fmt.Printf("Build    %s\n", build)
				fmt.Printf("ID       %v\n", jenkinstool.String(metadata.ID))
				fmt.Printf("Result   %v\n", jenkinstool.String(metadata.Result))

				for _, artifact := range metadata.Artifacts {
					fmt.Printf("Artifact %s\n", artifact.DisplayPath)
				}
				if showChanges {
					for _, cs := range metadata.ChangeSets {
						for _, item := range cs.Items {
							fmt.Printf("Change %s Comment %s\n", jenkinstool.String(item.Id), jenkinstool.String(item.Comment))
						}
					}
				}
			}

			if since == "" {
				return nil
			}

			if metadata.PreviousBuild == nil {
				return nil
			}

			prevBuild := jenkinstool.String(metadata.PreviousBuild.Number)
			if prevBuild == "" {
				return nil
			}

			build = prevBuild
			first = false
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
