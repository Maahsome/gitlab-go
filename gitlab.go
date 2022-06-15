package gitlab

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

type GitlabClient interface {
	GetProperty(property string) string
	SetProperty(property string, value string) string
	Get(uri string) (string, error)
	Delete(uri string) (string, error)
	GetUsers(search string) (string, error)
	GetGroup(groupID int) (Group, error)
	GetGroups(search string) (GroupList, error)
	GetSubGroups(groupID int) (GroupList, error)
	GetGroupProjects(groupID int) (ProjectList, error)
	GetGroupMembers(group int) (string, error)
	AddGroupMember(groupID, userID, accessLevel int) (string, error)
	GetForcePushSetting(projectID int, protectedBranch string) (bool, error)
	GetProjectID(projectPath string) (int, error)
	GetProject(projectID int) (Project, error)
	DeleteProject(projectID int) error
	GetProjectMirrors(projectID int) (ProjectMirrors, error)
	GetGroupID(groupPath string) (int, error)
	CreateProject(groupID int, projectPath string, visibility string) (Project, error)
	DeleteProtectedBranch(projectID int, protectedBranch string) (bool, error)
	ProtectBranch(projectID int, protectedBranch string) (bool, error)
	CreateProjectMirror(projectID int, mirrorURL string) (ProjectMirror, error)
	UpdateProjectMirror(projectID int, mirrorID int) (ProjectMirror, error)
	CreateMergeRequest(projectID int, title string, sourceBranch string, targetBranch string) (string, error)
}

type gitlabClient struct {
	BaseUrl      string
	ApiPath      string
	RepoFeedPath string
	Token        string
	Client       *resty.Client
}

// New generate a new gitlab client
func New(baseUrl, apiPath, token string) GitlabClient {

	// TODO: Add TLS Insecure && Pass in CA CRT for authentication
	restClient := resty.New()

	if apiPath == "" {
		apiPath = "/api/v4"
	}

	return &gitlabClient{
		BaseUrl: baseUrl,
		ApiPath: apiPath,
		Token:   token,
		Client:  restClient,
	}
}

func (r *gitlabClient) GetProperty(property string) string {
	switch property {
	case "BaseUrl":
		return r.BaseUrl
	case "ApiPath":
		return r.ApiPath
	case "Token":
		return r.Token
	}
	return ""
}

func (r *gitlabClient) SetProperty(property string, value string) string {
	switch property {
	case "BaseUrl":
		r.BaseUrl = value
		return r.BaseUrl
	case "ApiPath":
		r.ApiPath = value
		return r.ApiPath
	case "Token":
		r.Token = value
		return r.Token
	}
	return ""
}

func (r *gitlabClient) Get(uri string) (string, error) {

	nextPage := "1"
	combinedResults := ""

	for {
		// TODO: detect if there are no options passed in, ? verus & for page option
		fetchUri := fmt.Sprintf("https://%s%s%s&page=%s", r.BaseUrl, r.ApiPath, uri, nextPage)
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

func (r *gitlabClient) Delete(uri string) (string, error) {

	// https://git.alteryx.com/api/v4/projects/5784/releases/v0.0.6
	// GL_PAT=$(get-gitlab-api-pat)
	// curl --request DELETE --header "PRIVATE-TOKEN: ${GL_PAT}" "https://git.alteryx.com/api/v4/projects/5784/releases/v0.0.6"

	deleteUri := fmt.Sprintf("https://%s%s%s", r.BaseUrl, r.ApiPath, uri)
	// logrus.Warn(fetchUri)
	resp, resperr := r.Client.R().
		SetHeader("PRIVATE-TOKEN", r.Token).
		Delete(deleteUri)

	if resperr != nil {
		logrus.WithError(resperr).Error("Oops")
		return "", resperr
	}

	return string(resp.Body()[:]), nil
}

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
		logrus.Fatal("Cannot marshall Pipeline", marshErr)
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

// GetProjectMirrors - returns the full project based on the project ID
//
// GitLab API docs:
//
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

	// logrus.Info(fmt.Sprintf("%s", string(resp.Body()[:])))

	var prj Project
	marshErr := json.Unmarshal(resp.Body(), &prj)
	if marshErr != nil {
		logrus.Fatal("Cannot marshall Pipeline", marshErr)
		return Project{}, marshErr
	}

	return prj, nil

}

// ProtectBranch
//
// GitLab API docs:
//
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

// CreateProjectMirror creates a new mirror for a gitlab project (git repository)
//
// GitLab API docs:
//
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
//
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

// DeleteProtectedBranch - delete the specified branch from the protected list
//
// GitLab API docs:
//
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

// GetGroupID - returns the group ID based on the namespace/group path (slug)
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/groups.html#details-of-a-group
func (r *gitlabClient) GetGroupID(groupPath string) (int, error) {

	uri := fmt.Sprintf("/groups/%s", groupPath)
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
		logrus.Fatal("Cannot marshall Pipeline", marshErr)
		return 0, resperr
	}

	return pi.ID, nil

}

func (r *gitlabClient) GetUsers(search string) (string, error) {

	nextPage := "1"
	combinedResults := ""
	uri := "/users?active=true"
	for {
		// TODO: detect if there are no options passed in, ? verus & for page option
		fetchUri := fmt.Sprintf("https://%s%s%s&search=%s&page=%s", r.BaseUrl, r.ApiPath, uri, search, nextPage)
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

// GetGroup - returns the full group based on the group ID
//
// GitLab API docs:
//
func (r *gitlabClient) GetGroup(groupID int) (Group, error) {

	uri := fmt.Sprintf("/groups/%d", groupID)
	fetchUri := fmt.Sprintf("https://%s%s%s", r.BaseUrl, r.ApiPath, uri)
	// fmt.Printf("fetchUri: %s\n", fetchUri)
	resp, resperr := r.Client.R().
		SetHeader("PRIVATE-TOKEN", r.Token).
		SetHeader("Content-Type", "application/json").
		Get(fetchUri)

	if resperr != nil {
		logrus.WithError(resperr).Error("Oops")
		return Group{}, resperr
	}

	var gr Group
	marshErr := json.Unmarshal(resp.Body(), &gr)
	if marshErr != nil {
		logrus.Fatal("Cannot marshall Pipeline", marshErr)
		return Group{}, resperr
	}

	return gr, nil

}

// GetGroups- returns a list of groups
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/projects.html#get-single-project
func (r *gitlabClient) GetGroups(search string) (GroupList, error) {

	nextPage := "1"
	combinedResults := ""
	uri := "/groups?per_page=100&all_available=true"
	for {
		fetchUri := fmt.Sprintf("https://%s%s%s&page=%s", r.BaseUrl, r.ApiPath, uri, nextPage)
		// fmt.Printf("fetchUri: %s\n", fetchUri)
		resp, resperr := r.Client.R().
			SetHeader("PRIVATE-TOKEN", r.Token).
			SetHeader("Content-Type", "application/json").
			Get(fetchUri)

		if resperr != nil {
			logrus.WithError(resperr).Error("Oops")
			return GroupList{}, resperr
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
	surroundArray := fmt.Sprintf("[%s]", combinedResults)
	var gl GroupList
	marshErr := json.Unmarshal([]byte(surroundArray), &gl)
	if marshErr != nil {
		logrus.Fatal("Cannot marshall Pipeline", marshErr)
		return GroupList{}, marshErr
	}

	return gl, nil

}

// GetSubGroups- returns a list of subgroups for a given group id
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/groups.html#list-a-groups-subgroups
func (r *gitlabClient) GetSubGroups(groupID int) (GroupList, error) {

	nextPage := "1"
	combinedResults := ""
	uri := fmt.Sprintf("/groups/%d/subgroups", groupID)
	for {
		fetchUri := fmt.Sprintf("https://%s%s%s?page=%s", r.BaseUrl, r.ApiPath, uri, nextPage)
		// fmt.Printf("fetchUri: %s\n", fetchUri)
		resp, resperr := r.Client.R().
			SetHeader("PRIVATE-TOKEN", r.Token).
			SetHeader("Content-Type", "application/json").
			Get(fetchUri)

		if resperr != nil {
			logrus.WithError(resperr).Error("Oops")
			return GroupList{}, resperr
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
	surroundArray := fmt.Sprintf("[%s]", combinedResults)
	var gl GroupList
	marshErr := json.Unmarshal([]byte(surroundArray), &gl)
	if marshErr != nil {
		logrus.Fatal("Cannot marshall Pipeline", marshErr)
		return GroupList{}, marshErr
	}

	return gl, nil

}

// GetGroupProjects- returns a list of projects for a given group id
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/groups.html#list-a-groups-projects
func (r *gitlabClient) GetGroupProjects(groupID int) (ProjectList, error) {

	nextPage := "1"
	combinedResults := ""
	uri := fmt.Sprintf("/groups/%d/projects", groupID)
	for {
		fetchUri := fmt.Sprintf("https://%s%s%s?page=%s", r.BaseUrl, r.ApiPath, uri, nextPage)
		// fmt.Printf("fetchUri: %s\n", fetchUri)
		resp, resperr := r.Client.R().
			SetHeader("PRIVATE-TOKEN", r.Token).
			SetHeader("Content-Type", "application/json").
			Get(fetchUri)

		if resperr != nil {
			logrus.WithError(resperr).Error("Oops")
			return ProjectList{}, resperr
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
	surroundArray := fmt.Sprintf("[%s]", combinedResults)
	var pl ProjectList
	marshErr := json.Unmarshal([]byte(surroundArray), &pl)
	if marshErr != nil {
		logrus.Fatal("Cannot marshall Pipeline", marshErr)
		return ProjectList{}, marshErr
	}

	return pl, nil

}

func (r *gitlabClient) GetGroupMembers(group int) (string, error) {

	nextPage := "1"
	combinedResults := ""
	uri := fmt.Sprintf("/groups/%d/members", group)
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

// AddGroupMember
//
// GitLab API docs:
// https://docs.gitlab.com/ce/api/merge_requests.html#create-mr
func (r *gitlabClient) AddGroupMember(groupID, userID, accessLevel int) (string, error) {

	uri := fmt.Sprintf("/groups/%d/members", groupID)
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
