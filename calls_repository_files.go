package gitlab

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
)

// GetRepositoryFile - returns file contents from a repository file
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/repository_files.html#get-file-from-repository
func (r *gitlabClient) GetRepositoryFile(projectSlug string, fileSlug string, ref string) ([]byte, error) {

	// GET /projects/:id/repository/files/:file_path
	// curl --header "PRIVATE-TOKEN: ${GITLAB_TOKEN}" \
	//      --url "https://git.alteryx.com/api/v4/projects/futurama%2Fhermes%2Fcontrol-plane%2Fgcp%2Flowers%2Fc-us-e4-d00101/repository/files/%2Egitlab-ci%2Eyml?ref=master" \
	//      | jq -r '.content' | base64 -d

	uri := fmt.Sprintf("/projects/%s/repository/files/%s?ref=%s", projectSlug, fileSlug, ref)
	fetchUri := fmt.Sprintf("https://%s%s%s", r.BaseUrl, r.ApiPath, uri)
	// fmt.Printf("fetchUri: %s\n", fetchUri)
	resp, resperr := r.Client.R().
		SetHeader("PRIVATE-TOKEN", r.Token).
		SetHeader("Content-Type", "application/json").
		Get(fetchUri)

	if resperr != nil {
		logrus.WithError(resperr).Error("Oops")
		return []byte{}, resperr
	}

	var rf RepositoryFile
	marshErr := json.Unmarshal(resp.Body(), &rf)
	if marshErr != nil {
		return []byte{}, marshErr
	}

	fileBytes, err := base64.StdEncoding.DecodeString(rf.Content)
	if err != nil {
		return []byte{}, err
	}

	return fileBytes, nil
}
