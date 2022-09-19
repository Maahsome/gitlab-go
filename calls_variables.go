package gitlab

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

func getVariablesFrom(r *gitlabClient, id int, resource string) (Variables, error) {

	nextPage := "1"
	combinedResults := ""

	for {
		uri := fmt.Sprintf("/%s/%d/variables", resource, id)
		fetchUri := fmt.Sprintf("https://%s%s%s?page=%s", r.BaseUrl, r.ApiPath, uri, nextPage)
		resp, resperr := r.Client.R().
			SetHeader("PRIVATE-TOKEN", r.Token).
			SetHeader("Content-Type", "application/json").
			Get(fetchUri)

		if resperr != nil {
			logrus.WithError(resperr).Error("Oops")
			return Variables{}, resperr
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
	var variables Variables
	marshErr := json.Unmarshal([]byte(surroundArray), &variables)
	if marshErr != nil {
		logrus.Fatal("Cannot marshall Pipeline", marshErr)
		return Variables{}, marshErr
	}

	if resource == "projects" {
		projectInfo, perr := r.GetProject(id)
		if perr != nil {
			logrus.Error("Cannot marshall Pipeline", marshErr)
			return variables, perr
		}
		// for _, v := range variables {
		for k := range variables {
			// v.Source = projectInfo.Path
			variables[k].Source = projectInfo.PathWithNamespace
		}
	}

	if resource == "groups" {
		groupInfo, gerr := r.GetGroup(id)
		if gerr != nil {
			logrus.Error("Cannot marshall Pipeline", marshErr)
			return variables, gerr
		}
		for k := range variables {
			// variables[k].Source = groupInfo.Path
			variables[k].Source = groupInfo.FullPath
		}
	}
	return variables, nil
}

// GetCicdVariables - Returns all CICD Variables for a ProjectID
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/project_level_variables.html
// https://docs.gitlab.com/ee/api/group_level_variables.html
func (r *gitlabClient) GetCicdVariables(projectID int) (Variables, error) {

	// curl -Ls --header "PRIVATE-TOKEN: ${GITLAB_TOKEN}" "https://git.alteryx.com/api/v4/projects/5844/variables" | jq .

	// Fetch the ProjectID, extract .namespace.id (this is the immediate containing group)
	// Fetch the GroupID, extract .parent_id until 'null'

	uri := fmt.Sprintf("/projects/%d/variables", projectID)
	fetchUri := fmt.Sprintf("https://%s%s%s", r.BaseUrl, r.ApiPath, uri)
	resp, resperr := r.Client.R().
		SetHeader("PRIVATE-TOKEN", r.Token).
		SetHeader("Content-Type", "application/json").
		Get(fetchUri)

	if resperr != nil {
		logrus.WithError(resperr).Error("Oops")
		return Variables{}, resperr
	}

	var variables Variables
	marshErr := json.Unmarshal(resp.Body(), &variables)
	if marshErr != nil {
		logrus.Fatal("Cannot marshall Pipeline", marshErr)
		return Variables{}, resperr
	}

	projectInfo, perr := r.GetProject(projectID)
	if perr != nil {
		logrus.Error("Cannot marshall Pipeline", marshErr)
		return variables, perr
	}

	for k := range variables {
		variables[k].Source = projectInfo.PathWithNamespace
	}

	if projectInfo.Namespace.ID > 0 {
		groupID := projectInfo.Namespace.ID
		for {
			parentVariables, verr := getVariablesFrom(r, groupID, "groups")
			if verr != nil {
				logrus.Error("Cannot marshall Pipeline", marshErr)
				return variables, resperr
			}
			variables = append(variables, parentVariables...)
			groupInfo, gerr := r.GetGroup(groupID)
			if gerr != nil {
				logrus.Error("Cannot marshall Pipeline", marshErr)
				return variables, gerr
			}
			if groupInfo.ParentID == 0 {
				break
			}
			groupID = groupInfo.ParentID
		}
	}

	return variables, nil

}

// GetCicdVariablesFromGroup - Returns all CICD Variables for a ProjectID
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/project_level_variables.html
// https://docs.gitlab.com/ee/api/group_level_variables.html
func (r *gitlabClient) GetCicdVariablesFromGroup(groupID int, includeProjects bool) (Variables, error) {

	// curl -Ls --header "PRIVATE-TOKEN: ${GITLAB_TOKEN}" "https://git.alteryx.com/api/v4/projects/5844/variables" | jq .

	// Fetch the ProjectID, extract .namespace.id (this is the immediate containing group)
	// Fetch the GroupID, extract .parent_id until 'null'

	var variables Variables
	topVariables, verr := getVariablesFrom(r, groupID, "groups")
	if verr != nil {
		logrus.Error("Cannot marshall Pipeline", verr)
		return variables, verr
	}
	variables = append(variables, topVariables...)

	if includeProjects {
		topProjects, perr := r.GetGroupProjects(groupID)
		if perr != nil {
			logrus.Error("Failed to get top level group Projects", perr)
			return variables, perr
		}
		for _, v := range topProjects {
			projVariables, verr := getVariablesFrom(r, v.ID, "projects")
			if verr != nil {
				logrus.Error("Failed to get variables from topProjects ", verr)
				return variables, verr
			}
			variables = append(variables, projVariables...)
		}
	}
	subGroups, gerr := r.GetDescendantGroups(groupID)
	if gerr != nil {
		logrus.Error("Failed to get Top Project SubGroups", gerr)
		return variables, gerr
	}
	for _, v := range subGroups {
		grpVariables, verr := getVariablesFrom(r, v.ID, "groups")
		if verr != nil {
			logrus.Error("Failed to get variables from subGroups ", verr)
			return variables, verr
		}
		variables = append(variables, grpVariables...)
		if includeProjects {
			grpProjects, perr := r.GetGroupProjects(v.ID)
			if perr != nil {
				logrus.Error("Failed to get Projects for SubGroup ", perr)
				return variables, perr
			}
			for _, p := range grpProjects {
				projVariables, verr := getVariablesFrom(r, p.ID, "projects")
				if verr != nil {
					logrus.Error("Failed to get variables from topProjects ", verr)
					return variables, verr
				}
				variables = append(variables, projVariables...)
			}
		}

	}

	return variables, nil

}
