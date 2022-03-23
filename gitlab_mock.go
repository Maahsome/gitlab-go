package gitlab

import (
	"errors"
	"fmt"
	"strings"

	"github.com/stretchr/testify/mock"
)

type gitlabMock struct {
	BaseUrl      string
	ApiPath      string
	RepoFeedPath string
	Token        string
	Client       mock.Mock
}

type RequestError struct {
	StatusCode int

	Err error
}

func (r *RequestError) Error() string {
	return fmt.Sprintf("status %d: err %v", r.StatusCode, r.Err)
}

// NewGitlabMock - Mocking the gitlab interactions
func NewGitlabMock(baseUrl, apiPath, token string) GitlabClient {

	return &gitlabMock{
		BaseUrl: baseUrl,
		ApiPath: apiPath,
		Token:   token,
	}
}

func (r *gitlabMock) GetProperty(property string) string {
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

func (r *gitlabMock) SetProperty(property string, value string) string {
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

func (gm *gitlabMock) Get(uri string) (string, error) {

	if strings.Contains(uri, "error") {
		return "404 Not Found", &RequestError{
			StatusCode: 404,
			Err:        errors.New("not found"),
		}
	}

	return "some data", nil
}

func (gm *gitlabMock) Delete(uri string) (string, error) {

	// TODO: Return deletion status
	return "", nil
}

func (gm *gitlabMock) GetForcePushSetting(projectID int, protectedBranch string) (bool, error) {

	if protectedBranch == "error" {
		return false, &RequestError{
			StatusCode: 404,
			Err:        errors.New("not found"),
		}
	}
	return true, nil

}

func (gm *gitlabMock) GetProjectID(projectPath string) (int, error) {

	if strings.Contains(projectPath, "fail") {
		return 0, &RequestError{
			StatusCode: 404,
			Err:        errors.New("not found"),
		}
	}
	if strings.Contains(projectPath, "new") {
		return 0, nil
	}
	return 45, nil

}

func (gm *gitlabMock) GetProject(projectID int) (Project, error) {

	if projectID == 0 {
		return Project{}, &RequestError{
			StatusCode: 404,
			Err:        errors.New("not found"),
		}
	}
	return Project{}, nil
}

func (gm *gitlabMock) DeleteProject(projectID int) error {

	if projectID == 0 {
		return &RequestError{
			StatusCode: 404,
			Err:        errors.New("not found"),
		}
	}
	return nil
}

func (gm *gitlabMock) GetProjectMirrors(projectID int) (ProjectMirrors, error) {

	if projectID == 0 {
		return ProjectMirrors{}, &RequestError{
			StatusCode: 404,
			Err:        errors.New("not found"),
		}
	}
	return ProjectMirrors{}, nil
}

func (gm *gitlabMock) GetGroupID(groupPath string) (int, error) {
	if strings.Contains(groupPath, "error") {
		return 0, &RequestError{
			StatusCode: 404,
			Err:        errors.New("not found"),
		}
	}
	return 46, nil
}

func (gm *gitlabMock) GetUsers(search string) (string, error) {

	// TOOD: Add mock users return
	return "", nil
}

func (gm *gitlabMock) GetGroup(groupID int) (Group, error) {
	// TODO: return a sample group
	return Group{}, nil
}

func (gm *gitlabMock) GetGroups(search string) (GroupList, error) {
	// TODO: Add mock group list return
	return GroupList{}, nil
}

func (gm *gitlabMock) GetGroupMembers(group int) (string, error) {

	// TODO: Add mock group members return
	return "", nil
}

func (gm *gitlabMock) AddGroupMember(groupID, userID, accessLevel int) (string, error) {

	//TODO: Add mock return add user result
	return "", nil
}

func (gm *gitlabMock) CreateProject(groupID int, projectPath string, visibility string) (Project, error) {
	if visibility == "fail" {
		return Project{}, &RequestError{
			StatusCode: 404,
			Err:        errors.New("not found"),
		}
	}
	return Project{
		ID: 45,
	}, nil
}

func (gm *gitlabMock) DeleteProtectedBranch(projectID int, protectedBranch string) (bool, error) {
	if strings.Contains(protectedBranch, "error") {
		return false, &RequestError{
			StatusCode: 404,
			Err:        errors.New("not found"),
		}
	}
	return true, nil
}

func (gm *gitlabMock) ProtectBranch(projectID int, protectedBranch string) (bool, error) {
	if strings.Contains(protectedBranch, "error") {
		return false, &RequestError{
			StatusCode: 404,
			Err:        errors.New("not found"),
		}
	}
	return true, nil
}

func (gm *gitlabMock) CreateProjectMirror(projectID int, mirrorURL string) (ProjectMirror, error) {
	if strings.Contains(mirrorURL, "fail") {
		return ProjectMirror{}, &RequestError{
			StatusCode: 404,
			Err:        errors.New("not found"),
		}
	}
	return ProjectMirror{
		Enabled:               true,
		OnlyProtectedBranches: true,
	}, nil
}

func (gm *gitlabMock) UpdateProjectMirror(projectID int, mirrorID int) (ProjectMirror, error) {
	if projectID == 0 {
		return ProjectMirror{}, &RequestError{
			StatusCode: 404,
			Err:        errors.New("not found"),
		}
	}
	return ProjectMirror{}, nil
}

func (gm *gitlabMock) CreateMergeRequest(projectID int, title string, sourceBranch string, targetBranch string) (string, error) {
	if projectID == 0 {
		return "", &RequestError{
			StatusCode: 404,
			Err:        errors.New("not found"),
		}
	}
	return "", nil
}
