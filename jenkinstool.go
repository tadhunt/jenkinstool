package main

import(
	"fmt"
	"io"
	"net/url"
	"net/http"
	"encoding/json"
	"github.com/integrii/flaggy"
)

type Cmd struct {
	cmd     *flaggy.Subcommand
	handler func(cmd *Cmd) error
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
	flaggy.String(&server, "server", "", "URL of Jenkins server to interact with")

	cmds := []*Cmd{
		newGetCmd(),
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
		if (cmd.cmd.Used) {
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

	get.String(&build, "build", "", "Build to fetch (defaults to latest)")

	handler := func(cmd *Cmd) error {
		switch build {
		case "":
			fallthrough
		case "latest":
			build = "lastSuccessfulBuild"
		}

		u := fmt.Sprintf("%s/%s/api/json", serverURL.String(), build)

		fmt.Printf("GET %s\n", u)

		response, err := http.Get(u)
		if err != nil {
			return err
		}

		body, err := io.ReadAll(response.Body)
		if err != nil {
			return err
		}

		result := make(map[string]any)
		err = json.Unmarshal(body, &result)
		if err != nil {
			return err
		}

		fmt.Printf("Build %s\n", build)
		fmt.Printf("ID %v\n", result["id"])

		return nil
	}

	return &Cmd{cmd: get, handler: handler}
}
