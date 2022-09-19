package gitlab

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

// GetPipelines returns a list of pipelines for the project
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/pipelines.html#list-project-pipelines
func (r *gitlabClient) GetPipelines(projectID int, user string) (Pipelines, error) {

	nextPage := "1"
	combinedResults := ""
	uri := ""
	if len(user) > 0 {
		uri = fmt.Sprintf("/projects/%d/pipelines?per_page=100&username=%s", projectID, user)
	} else {
		uri = fmt.Sprintf("/projects/%d/pipelines?per_page=100", projectID)
	}
	for {
		fetchUri := fmt.Sprintf("https://%s%s%s&page=%s", r.BaseUrl, r.ApiPath, uri, nextPage)
		// fmt.Printf("fetchUri: %s\n", fetchUri)
		resp, resperr := r.Client.R().
			SetHeader("PRIVATE-TOKEN", r.Token).
			SetHeader("Content-Type", "application/json").
			Get(fetchUri)

		if resperr != nil {
			logrus.WithError(resperr).Error("Oops")
			return Pipelines{}, resperr
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
	var pipelines Pipelines
	marshErr := json.Unmarshal([]byte(surroundArray), &pipelines)
	if marshErr != nil {
		logrus.Fatal("Cannot marshall Pipeline", marshErr)
		return Pipelines{}, marshErr
	}

	return pipelines, nil

}

// GetPipeline - Returns a single pipeline
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/pipelines.html#get-a-single-pipeline
func (r *gitlabClient) GetPipeline(projectID int, pipelineID int) (Pipeline, error) {

	uri := fmt.Sprintf("/projects/%d/pipelines/%d", projectID, pipelineID)
	fetchUri := fmt.Sprintf("https://%s%s%s", r.BaseUrl, r.ApiPath, uri)
	// fmt.Printf("fetchUri: %s\n", fetchUri)
	resp, resperr := r.Client.R().
		SetHeader("PRIVATE-TOKEN", r.Token).
		SetHeader("Content-Type", "application/json").
		Get(fetchUri)

	if resperr != nil {
		logrus.WithError(resperr).Error("Oops")
		return Pipeline{}, resperr
	}

	var pipeline Pipeline
	marshErr := json.Unmarshal(resp.Body(), &pipeline)
	if marshErr != nil {
		logrus.Fatal("Cannot marshall Pipeline", marshErr)
		return Pipeline{}, resperr
	}

	return pipeline, nil

}
