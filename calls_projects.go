package gitlab

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

// GetProjectID - returns the project ID based on the group/project path (slug)
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/projects.html#get-single-project
func (r *gitlabClient) GetProjectID(projectPath string) (int, error) {

	uri := fmt.Sprintf("/projects/%s", projectPath)
	fetchUri := fmt.Sprintf("https://%s%s%s", r.BaseUrl, r.ApiPath, uri)
	// fmt.Printf("fetchUri: %s\n", fetchUri)
	resp, resperr := r.Client.R().
		SetHeader("PRIVATE-TOKEN", r.Token).
		SetHeader("Content-Type", "application/json").
		Get(fetchUri)

	if resperr != nil {
		logrus.WithError(resperr).Error("Oops")
		return 0, resperr
	}

	var pi ProjectInfo
	marshErr := json.Unmarshal(resp.Body(), &pi)
	if marshErr != nil {
		return 0, errors.New(fmt.Sprintf("Error unmarshalling Project Response, %e", marshErr))
	}

	return pi.ID, nil

}

// GetProject - returns the full project based on the project ID
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/projects.html#get-single-project
func (r *gitlabClient) GetProject(projectID int) (Project, error) {

	uri := fmt.Sprintf("/projects/%d", projectID)
	fetchUri := fmt.Sprintf("https://%s%s%s", r.BaseUrl, r.ApiPath, uri)
	// fmt.Printf("fetchUri: %s\n", fetchUri)
	resp, resperr := r.Client.R().
		SetHeader("PRIVATE-TOKEN", r.Token).
		SetHeader("Content-Type", "application/json").
		Get(fetchUri)

	if resperr != nil {
		logrus.WithError(resperr).Error("Oops")
		return Project{}, resperr
	}

	var pi Project
	marshErr := json.Unmarshal(resp.Body(), &pi)
	if marshErr != nil {
		logrus.Fatal("Cannot marshall Pipeline", marshErr)
		return Project{}, resperr
	}

	return pi, nil

}

// CreateProject creates a new gitlab project (git repository)
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/projects.html#create-project
func (r *gitlabClient) CreateProject(groupID int, projectPath string, visibility string) (Project, error) {
	// curl -Ls --request POST https://gitlab.com/api/v4/projects --header "PRIVATE-TOKEN: ${mytoken}" \
	//  --header 'Content-Type: application/json' \
	//  --data "{
	//            \"path\": \"${new_ring}\",
	//            \"default_branch\": \"master\",
	//            \"initialize_with_readme\": \"true\",
	//            \"visibility\": \"private\",
	//            \"namespace_id\": \"${group_id}\"
	//   }"

	uri := "/projects"
	fetchUri := fmt.Sprintf("https://%s%s%s", r.BaseUrl, r.ApiPath, uri)
	// logrus.Info(fmt.Sprintf("fetchUri: %s", fetchUri))
	projectTemplate := `{
			"path": "%s",
			"default_branch": "master",
			"initialize_with_readme": "true",
			"visibility": "%s",
			"namespace_id": "%d"
			}`
	body := fmt.Sprintf(projectTemplate, projectPath, visibility, groupID)
	// logrus.Info(fmt.Sprintf("body: %s", body))
	resp, resperr := r.Client.R().
		SetHeader("PRIVATE-TOKEN", r.Token).
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post(fetchUri)

	if resperr != nil {
		logrus.WithError(resperr).Error("Oops")
		return Project{}, resperr
	}

	logrus.Info(fmt.Sprintf("%s", string(resp.Body()[:])))

	var prj Project
	marshErr := json.Unmarshal(resp.Body(), &prj)
	if marshErr != nil {
		logrus.Fatal("Cannot marshall Pipeline", marshErr)
		return Project{}, marshErr
	}

	return prj, nil

}

// DeleteProject - Delete the project by ProjectID
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/projects.html#delete-project
func (r *gitlabClient) DeleteProject(projectID int) error {

	uri := fmt.Sprintf("/projects/%d", projectID)
	fetchUri := fmt.Sprintf("https://%s%s%s", r.BaseUrl, r.ApiPath, uri)
	resp, resperr := r.Client.R().
		SetHeader("PRIVATE-TOKEN", r.Token).
		SetHeader("Content-Type", "application/json").
		Delete(fetchUri)

	if resperr != nil {
		return resperr
	}

	var msg Message
	marshErr := json.Unmarshal(resp.Body(), &msg)
	if marshErr != nil {
		return marshErr
	}

	if msg.Message == "202 Accepted" {
		return nil
	}
	return errors.New(strings.ToLower(msg.Message))

}

// https://docs.gitlab.com/ee/api/members.html#list-all-members-of-a-group-or-project
func (r *gitlabClient) GetProjectMembers(project int) (string, error) {

	nextPage := "1"
	combinedResults := ""
	uri := fmt.Sprintf("/projects/%d/members", project)
	for {
		// TODO: detect if there are no options passed in, ? verus & for page option
		fetchUri := fmt.Sprintf("https://%s%s%s?page=%s", r.BaseUrl, r.ApiPath, uri, nextPage)
		// logrus.Warn(fetchUri)
		resp, resperr := r.Client.R().
			SetHeader("PRIVATE-TOKEN", r.Token).
			Get(fetchUri)

		if resperr != nil {
			logrus.WithError(resperr).Error("Oops")
			return "", resperr
		}

		items := strings.TrimPrefix(string(resp.Body()[:]), "[")
		items = strings.TrimSuffix(items, "]")
		if combinedResults == "" {
			combinedResults += items
		} else {
			combinedResults += fmt.Sprintf(", %s", items)
		}
		currentPage := resp.Header().Get("X-Page")
		nextPage = resp.Header().Get("X-Next-Page")
		totalPages := resp.Header().Get("X-Total-Pages")
		if currentPage == totalPages {
			break
		}
	}
	return fmt.Sprintf("[%s]", combinedResults), nil
}

// AddProjectMember
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/members.html#add-a-member-to-a-group-or-project
func (r *gitlabClient) AddProjectMember(projectID, userID, accessLevel int) (string, error) {

	uri := fmt.Sprintf("/projects/%d/members", projectID)
	fetchUri := fmt.Sprintf("https://%s%s%s", r.BaseUrl, r.ApiPath, uri)
	memberTemplate := `{
			"user_id": "%d",
			"access_level": "%d"
			}`
	body := fmt.Sprintf(memberTemplate, userID, accessLevel)
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

// GetProjectMirrors - returns the full project based on the project ID
//
// GitLab API docs:
func (r *gitlabClient) GetProjectMirrors(projectID int) (ProjectMirrors, error) {

	// curl -Ls "https://git.alteryx.com/api/v4/projects/${PR_ID}/remote_mirrors" \
	//     --header "PRIVATE-TOKEN: ${PRIV_TOKEN}" | jq -r '.[] | ([.id,.url,.enabled,.only_protected_branches]

	uri := fmt.Sprintf("/projects/%d/remote_mirrors", projectID)
	fetchUri := fmt.Sprintf("https://%s%s%s", r.BaseUrl, r.ApiPath, uri)
	// fmt.Printf("fetchUri: %s\n", fetchUri)
	resp, resperr := r.Client.R().
		SetHeader("PRIVATE-TOKEN", r.Token).
		SetHeader("Content-Type", "application/json").
		Get(fetchUri)

	if resperr != nil {
		logrus.WithError(resperr).Error("Oops")
		return ProjectMirrors{}, resperr
	}

	var prm ProjectMirrors
	marshErr := json.Unmarshal(resp.Body(), &prm)
	if marshErr != nil {
		logrus.Fatal("Cannot marshall Pipeline", marshErr)
		return ProjectMirrors{}, resperr
	}

	return prm, nil

}

// ProtectBranch
//
// GitLab API docs:
func (r *gitlabClient) ProtectBranch(projectID int, protectedBranch string) (bool, error) {

	// curl -Ls --request POST "https://gitlab.com/api/v4/projects/${PR_ID}/protected_branches" \
	//          --header "PRIVATE-TOKEN: ${PUB_TOKEN}" \
	//          --header 'Content-Type: application/json' \
	//          --data "{
	//                  \"name\": \"master\",
	//                  \"push_access_levels\": [
	//                    {
	//                      \"access_level\": 40,
	//                      \"access_level_description\": \"Maintainers\",
	//                      \"user_id\": null,
	//                      \"group_id\": null
	//                    }
	//                  ],
	//                  \"merge_access_levels\": [
	//                    {
	//                      \"access_level\": 40,
	//                      \"access_level_description\": \"Maintainers\",
	//                      \"user_id\": null,
	//                      \"group_id\": null
	//                    }
	//                  ],
	//                  \"allow_force_push\": true,
	//                  \"unprotect_access_levels\": [],
	//                  \"code_owner_approval_required\": false
	//                }" | jq '.allow_force_push'

	uri := fmt.Sprintf("/projects/%d/protected_branches", projectID)
	fetchUri := fmt.Sprintf("https://%s%s%s", r.BaseUrl, r.ApiPath, uri)
	pbTemplate := `{
			"name": "%s",
			"push_access_levels": [
				{
					"access_level": 40,
					"access_level_description": "Maintainers",
					"user_id": null,
					"group_id": null
				}
			],
			"merge_access_levels": [
				{
					"access_level": 40,
					"access_level_description": "Maintainers",
					"user_id": null,
					"group_id": null
				}
			],
			"allow_force_push": true,
			"unprotect_access_levels": [],
			"code_owner_approval_required": false
			}`
	body := fmt.Sprintf(pbTemplate, protectedBranch)
	resp, resperr := r.Client.R().
		SetHeader("PRIVATE-TOKEN", r.Token).
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post(fetchUri)

	if resperr != nil {
		logrus.WithError(resperr).Error("Oops")
		return false, resperr
	}

	var pbs ProtectedBranchSettings
	marshErr := json.Unmarshal(resp.Body(), &pbs)
	if marshErr != nil {
		logrus.Fatal("Cannot marshall Pipeline", marshErr)
		return false, resperr
	}

	return pbs.AllowForcePush, nil

}

// DeleteProtectedBranch - delete the specified branch from the protected list
//
// GitLab API docs:
func (r *gitlabClient) DeleteProtectedBranch(projectID int, protectedBranch string) (bool, error) {

	// curl -Ls --request DELETE "https://gitlab.com/api/v4/projects/${PR_ID}/protected_branches/master" \
	//      --header "PRIVATE-TOKEN: ${PUB_TOKEN}" | jq -r '.message'
	uri := fmt.Sprintf("/projects/%d/protected_branches/%s", projectID, protectedBranch)
	fetchUri := fmt.Sprintf("https://%s%s%s", r.BaseUrl, r.ApiPath, uri)
	// fmt.Printf("fetchUri: %s\n", fetchUri)
	resp, resperr := r.Client.R().
		SetHeader("PRIVATE-TOKEN", r.Token).
		SetHeader("Content-Type", "application/json").
		Delete(fetchUri)

	if resperr != nil {
		logrus.WithError(resperr).Error("Oops")
		return false, resperr
	}

	return resp.IsSuccess(), nil
}

// CreateProjectMirror creates a new mirror for a gitlab project (git repository)
//
// GitLab API docs:
func (r *gitlabClient) CreateProjectMirror(projectID int, mirrorURL string) (ProjectMirror, error) {

	// curl -Ls --request POST "https://git.alteryx.com/api/v4/projects/${PR_ID}/remote_mirrors" \
	//     --header "PRIVATE-TOKEN: ${PRIV_TOKEN}" \
	//     --header 'Content-Type: application/json' \
	//     --data "{
	//       \"url\": \"https://falkor_sync:${ENC_SYNC_PW}@gitlab.com/${PROJECT_PATH}\",
	//       \"enabled\": \"true\",
	//       \"only_protected_branches\": \"true\"
	//     }" | jq -r '([.id,.url,.enabled,.only_protected_branches]) | @csv')

	uri := fmt.Sprintf("/projects/%d/remote_mirrors", projectID)
	fetchUri := fmt.Sprintf("https://%s%s%s", r.BaseUrl, r.ApiPath, uri)
	// logrus.Info(fmt.Sprintf("fetchUri: %s", fetchUri))
	mirrorTemplate := `{
			"url": "%s",
			"enabled": "true",
			"only_protected_branches": "true"
			}`
	body := fmt.Sprintf(mirrorTemplate, mirrorURL)
	// logrus.Info(fmt.Sprintf("body: %s", body))
	resp, resperr := r.Client.R().
		SetHeader("PRIVATE-TOKEN", r.Token).
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post(fetchUri)

	if resperr != nil {
		logrus.WithError(resperr).Error("Oops")
		return ProjectMirror{}, resperr
	}

	// logrus.Info(fmt.Sprintf("%s", string(resp.Body()[:])))

	var pm ProjectMirror
	marshErr := json.Unmarshal(resp.Body(), &pm)
	if marshErr != nil {
		logrus.Fatal("Cannot marshall Pipeline", marshErr)
		return ProjectMirror{}, resperr
	}

	return pm, nil

}

// UpdateProjectMirror update a project mirror settings for a gitlab project (git repository)
//
// GitLab API docs:
func (r *gitlabClient) UpdateProjectMirror(projectID int, mirrorID int) (ProjectMirror, error) {

	// curl -Ls --request PUT https://git.alteryx.com/api/v4/projects/${PR_ID}/remote_mirrors/${m_id} \
	//             --header "PRIVATE-TOKEN: ${PRIV_TOKEN}" \
	//             --header 'Content-Type: application/json' \
	//             --data "{
	//               \"enabled\": \"true\",
	//               \"only_protected_branches\": \"true\"
	//             }" | jq -r '([.id,.url,.enabled,.only_protected_branches]) | @csv')

	uri := fmt.Sprintf("/projects/%d/remote_mirrors/%d", projectID, mirrorID)
	fetchUri := fmt.Sprintf("https://%s%s%s", r.BaseUrl, r.ApiPath, uri)
	// logrus.Info(fmt.Sprintf("fetchUri: %s", fetchUri))
	mirrorTemplate := "enabled=true&only_protected_branches=true"
	resp, resperr := r.Client.R().
		SetHeader("PRIVATE-TOKEN", r.Token).
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetBody(mirrorTemplate).
		Put(fetchUri)

	if resperr != nil {
		logrus.WithError(resperr).Error("Oops")
		return ProjectMirror{}, resperr
	}

	// logrus.Info(fmt.Sprintf("%s", string(resp.Body()[:])))

	var pm ProjectMirror
	marshErr := json.Unmarshal(resp.Body(), &pm)
	if marshErr != nil {
		logrus.Fatal("Cannot marshall Pipeline", marshErr)
		return ProjectMirror{}, resperr
	}

	return pm, nil

}

// GetForcePushSetting -
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/projects.html#get-single-project
func (r *gitlabClient) GetForcePushSetting(projectID int, protectedBranch string) (bool, error) {

	// curl -Ls --header "PRIVATE-TOKEN: ${PUB_TOKEN}" "https://gitlab.com/api/v4/projects/${PR_ID}/protected_branches/master" | jq '.allow_force_push'
	uri := fmt.Sprintf("/projects/%d/protected_branches/%s", projectID, protectedBranch)
	fetchUri := fmt.Sprintf("https://%s%s%s", r.BaseUrl, r.ApiPath, uri)
	// fmt.Printf("fetchUri: %s\n", fetchUri)
	resp, resperr := r.Client.R().
		SetHeader("PRIVATE-TOKEN", r.Token).
		SetHeader("Content-Type", "application/json").
		Get(fetchUri)

	if resperr != nil {
		logrus.WithError(resperr).Error("Oops")
		return false, resperr
	}

	var pbs ProtectedBranchSettings
	marshErr := json.Unmarshal(resp.Body(), &pbs)
	if marshErr != nil {
		logrus.Fatal("Cannot marshall Pipeline", marshErr)
		return false, resperr
	}

	return pbs.AllowForcePush, nil

}
