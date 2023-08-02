package jenkinstool

const (
	LatestBuild = "latest"
)

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
