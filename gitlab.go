package giblab

import (
	"encoding/json"
	"flag"
	"fmt"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

type Gitlab struct {
	BaseUrl      string
	ApiPath      string
	RepoFeedPath string
	Token        string
	Client       *resty.Client
}

type PaginationOptions struct {
	Page    int `url:"page,omitempty"`
	PerPage int `url:"per_page,omitempty"`
}

type SortDirection string

const (
	SortDirectionAsc  SortDirection = "asc"
	SortDirectionDesc SortDirection = "desc"
)

type SortOptions struct {
	OrderBy string        `url:"order_by,omitempty"`
	Sort    SortDirection `url:"sort,omitempty"`
}

type ResponseWithMessage struct {
	Message string `json:"message"`
}

type ResponseMeta struct {
	Method     string
	Url        string
	StatusCode int
	RequestId  string
	Page       int
	PerPage    int
	PrevPage   int
	NextPage   int
	TotalPages int
	Total      int
	Runtime    float64
}

type ProjectInfo struct {
	ID int `json:"id"`
}

const (
	dateLayout = "2006-01-02T15:04:05-07:00"
)

var (
	skipCertVerify = flag.Bool("gitlab.skip-cert-check", false,
		`If set to true, gitlab client will skip certificate checking for https, possibly exposing your system to MITM attack.`)
)

// New generate a new gitlab client
func New(baseUrl, apiPath, token string) *Gitlab {

	// TODO: Add TLS Insecure && Pass in CA CRT for authentication
	restClient := resty.New()

	if apiPath == "" {
		apiPath = "/api/v4"
	}

	return &Gitlab{
		BaseUrl: baseUrl,
		ApiPath: apiPath,
		Token:   token,
		Client:  restClient,
	}
}

func (r *Gitlab) Get(uri string) (string, error) {

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

func (r *Gitlab) Delete(uri string) (string, error) {

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
func (r *Gitlab) CreateMergeRequest(projectID int, title string, sourceBranch string, targetBranch string) (string, error) {
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
func (r *Gitlab) GetProjectID(projectPath string) (int, error) {

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

func (r *Gitlab) GetUsers(search string) (string, error) {

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

func (r *Gitlab) GetGroupMembers(group int) (string, error) {

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
func (r *Gitlab) AddGroupMember(groupID, userID, accessLevel int) (string, error) {

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
