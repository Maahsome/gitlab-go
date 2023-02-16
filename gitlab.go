package gitlab

import (
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
	GetDescendantGroups(groupID int) (GroupList, error)
	GetGroupProjects(groupID int) (ProjectList, error)
	GetGroupMembers(group int) (string, error)
	AddGroupMember(groupID, userID, accessLevel int) (string, error)
	GetForcePushSetting(projectID int, protectedBranch string) (bool, error)
	GetProjectID(projectPath string) (int, error)
	GetProject(projectID int) (Project, error)
	GetProjectMembers(project int) (string, error)
	AddProjectMember(projectID, userID, accessLevel int) (string, error)
	DeleteProject(projectID int) error
	GetProjectMirrors(projectID int) (ProjectMirrors, error)
	GetGroupID(groupPath string) (int, error)
	CreateProject(groupID int, projectPath string, visibility string) (Project, error)
	DeleteProtectedBranch(projectID int, protectedBranch string) (bool, error)
	ProtectBranch(projectID int, protectedBranch string) (bool, error)
	CreateProjectMirror(projectID int, mirrorURL string) (ProjectMirror, error)
	UpdateProjectMirror(projectID int, mirrorID int) (ProjectMirror, error)
	CreateMergeRequest(projectID int, title string, sourceBranch string, targetBranch string) (string, error)
	GetPipelines(projectID int, user string, limit int) (Pipelines, error)
	GetPipeline(projectID int, pipelineID int) (Pipeline, error)
	GetVariableFrom(id int, resource string, variable string) (string, error)
	GetCicdVariables(projectdID int) (Variables, error)
	GetCicdVariablesFromGroup(groupID int, includeProjects bool) (Variables, error)
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
