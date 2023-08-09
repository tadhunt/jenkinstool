package jenkinstool

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type BuildMetadata struct {
	ID            *string      `json:"id"`
	Result        *string      `json:"result"`
	Artifacts     []*Artifact  `json:"artifacts"`
	ChangeSets    []*ChangeSet `json:"changeSets"`
	InProgress    bool         `json:"inProgress"`
	NextBuild     *BuildInfo   `json:"nextBuild"`
	PreviousBuild *BuildInfo   `json:"previousBuild"`
}

type Artifact struct {
	DisplayPath  string `json:"displayPath"`
	Filename     string `json:"fileName"`
	RelativePath string `json:"relativePath"`
}

/*
   "_class": "hudson.plugins.git.GitChangeSetList",
   "items": [
     {
       "_class": "hudson.plugins.git.GitChangeSet",
       "affectedPaths": [
         "core/src/main/java/org/geysermc/geyser/Constants.java",
         "core/src/main/java/org/geysermc/geyser/command/defaults/VersionCommand.java"
       ],
       "commitId": "505c06595633393601bf8c8127d3368172a43096",
       "timestamp": 1691515123000,
       "author": {
         "absoluteUrl": "https://ci.opencollab.dev/user/github",
         "fullName": "github"
       },
       "authorEmail": "noreply@github.com",
       "comment": "Update Geyser download URL (#4045)\n\n* Update Geyser download URL\n\n* Use existing constant instead of duplicating string\n",
       "date": "2023-08-08 10:18:43 -0700",
       "id": "505c06595633393601bf8c8127d3368172a43096",
       "msg": "Update Geyser download URL (#4045)",
       "paths": [
         {
           "editType": "edit",
           "file": "core/src/main/java/org/geysermc/geyser/Constants.java"
         },
         {
           "editType": "edit",
           "file": "core/src/main/java/org/geysermc/geyser/command/defaults/VersionCommand.java"
         }
       ]
     }
   ]
*/

type ChangeSet struct {
	Class *string          `json:"_class"`
	Items []*ChangeSetItem `json:"items"`
}

type ChangeSetItem struct {
	Class         *string          `json:"_class"`
	AffectedPaths []string         `json:"affectedPaths"`
	CommitId      *string          `json:"commitId"`
	Timestamp     *JsonTime        `json:"timestamp"`
	Author        *ChangeSetAuthor `json:"author"`
	AuthorEmail   *string          `json:"authorEmail"`
	Comment       *string          `json:"comment"`
	Date          *string          `json:"date"`
	Id            *string          `json:"id"`
	Msg           *string          `json:"msg"`
	Paths         []*ChangeSetPath `json:"paths"`
}

type ChangeSetAuthor struct {
	AbsoluteURL *string `json:"absoluteUrl"`
	FullName    *string `json:"fullName"`
}

type ChangeSetPath struct {
	EditType *string `json:"editType"`
	File     *string `json:"file"`
}

type BuildInfo struct {
	Number *float64 `json:"number"`
	URL    *string  `json:"url"`
}

type MetadataSyntaxError struct {
	Raw    string
	msg    string
	Offset int64
}

func (e *MetadataSyntaxError) Error() string {
	return e.msg
}

func GetBuildMetadata(src *url.URL, build string) (*BuildMetadata, error) {
	build = parseBuild(build)

	u := fmt.Sprintf("%s/%s/api/json", src.String(), build)

	response, err := http.Get(u)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return GetBuildMetadataFromBytes(body)
}

func GetRawBuildMetadata(src *url.URL, build string) ([]byte, error) {
	build = parseBuild(build)

	u := fmt.Sprintf("%s/%s/api/json", src.String(), build)

	response, err := http.Get(u)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func GetBuildMetadataFromBytes(raw []byte) (*BuildMetadata, error) {
	metadata := &BuildMetadata{}
	err := json.Unmarshal(raw, metadata)
	if err != nil {
		serr, isSyntaxError := err.(*json.SyntaxError)
		if isSyntaxError {
			return nil, &MetadataSyntaxError{
				Raw:    string(raw),
				msg:    fmt.Sprintf("%v (offset %d)", serr, serr.Offset),
				Offset: serr.Offset,
			}
		}
		return nil, err
	}

	return metadata, nil
}
