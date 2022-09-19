package gitlab

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

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

// GetGroup - returns the full group based on the group ID
//
// GitLab API docs:
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

// GetDescendantGroups - returns a list of descendant_groups for a groupID
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/groups.html#list-a-groups-descendant-groups
func (r *gitlabClient) GetDescendantGroups(groupID int) (GroupList, error) {

	nextPage := "1"
	combinedResults := ""
	uri := fmt.Sprintf("/groups/%d/descendant_groups", groupID)
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

// https://docs.gitlab.com/ee/api/members.html#list-all-members-of-a-group-or-project
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
// https://docs.gitlab.com/ee/api/members.html#add-a-member-to-a-group-or-project
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
