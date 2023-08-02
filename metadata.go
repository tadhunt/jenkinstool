package jenkinstool

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
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

func GetBuildMetadata(src *url.URL, build string) (*BuildMetadata, error) {
	u := fmt.Sprintf("%s/%s/api/json", src.String(), build)

	response, err := http.Get(u)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	metadata := &BuildMetadata{}
	err = json.Unmarshal(body, metadata)
	if err != nil {
		return nil, err
	}

	return metadata, nil
}

func GetRawBuildMetadata(src *url.URL, build string) (string, error) {
	u := fmt.Sprintf("%s/%s/api/json", src.String(), build)

	response, err := http.Get(u)
	if err != nil {
		return "", err
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
