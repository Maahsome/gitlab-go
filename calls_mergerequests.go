package gitlab

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

// CreateMergeRequest creates a new merge request.
//
// GitLab API docs:
// https://docs.gitlab.com/ce/api/merge_requests.html#create-mr
func (r *gitlabClient) CreateMergeRequest(projectID int, title string, sourceBranch string, targetBranch string) (string, error) {
	//                      https://git.alteryx.com/api/v4/projects/5701         /merge_requests
	// 	curl --request POST https://gitlab.com     /api/v4/projects/${project_id}/merge_requests --header "PRIVATE-TOKEN: ${mytoken}" \
	//   --header 'Content-Type: application/json' \
	//   --data "{
	//             \"id\": \"${project_id}\",
	//             \"title\": \"m2d\",
	//             \"source_branch\": \"m2d\",
	//             \"target_branch\": \"develop\"
	//     }"

	uri := fmt.Sprintf("/projects/%d/merge_requests", projectID)
	fetchUri := fmt.Sprintf("https://%s%s%s", r.BaseUrl, r.ApiPath, uri)
	mrTemplate := `{
			"id": "%d",
			"title": "%s",
			"source_branch": "%s",
			"target_branch": "%s"
			}`
	body := fmt.Sprintf(mrTemplate, projectID, title, sourceBranch, targetBranch)
	resp, resperr := r.Client.R().
		SetHeader("PRIVATE-TOKEN", r.Token).
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post(fetchUri)

	if resperr != nil {
		logrus.WithError(resperr).Error("Oops")
		return "", resperr
	}

	return string(resp.Body()[:]), nil

}
