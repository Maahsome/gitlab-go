package gitlab

import "time"

type ProjectInfo struct {
	ID int `json:"id"`
}

type ProjectList []Project

type Project struct {
	ID                int       `json:"id"`
	Name              string    `json:"name"`
	NameWithNamespace string    `json:"name_with_namespace"`
	Path              string    `json:"path"`
	PathWithNamespace string    `json:"path_with_namespace"`
	Archived          bool      `json:"archived"`
	SSHURL            string    `json:"ssh_url_to_repo"`
	HTTPURL           string    `json:"http_url_to_repo"`
	CreatedAt         time.Time `json:"created_at"`
	DefaultBranch     string    `json:"default_branch"`
	Namespace         struct {
		ID       int    `json:"id"`
		Name     string `json:"name"`
		Path     string `json:"path"`
		Kind     string `json:"kind"`
		FullPath string `json:"full_path"`
		ParentID int    `json:"parent_id"`
		WebURL   string `json:"web_url"`
	} `json:"namespace"`
	Visibility string `json:"visibility"`
	CreatorID  int    `json:"creator_id"`
	Mirror     bool   `json:"mirror"`
}

type ProtectedBranchSettings struct {
	ID               int    `json:"id"`
	Name             string `json:"name"`
	PushAccessLevels []struct {
		AccessLevel            int         `json:"access_level"`
		AccessLevelDescription string      `json:"access_level_description"`
		UserID                 interface{} `json:"user_id"`
		GroupID                interface{} `json:"group_id"`
	} `json:"push_access_levels"`
	MergeAccessLevels []struct {
		AccessLevel            int         `json:"access_level"`
		AccessLevelDescription string      `json:"access_level_description"`
		UserID                 interface{} `json:"user_id"`
		GroupID                interface{} `json:"group_id"`
	} `json:"merge_access_levels"`
	AllowForcePush        bool `json:"allow_force_push"`
	UnprotectAccessLevels []struct {
		AccessLevel            int         `json:"access_level"`
		AccessLevelDescription string      `json:"access_level_description"`
		UserID                 interface{} `json:"user_id"`
		GroupID                interface{} `json:"group_id"`
	} `json:"unprotect_access_levels"`
	CodeOwnerApprovalRequired bool `json:"code_owner_approval_required"`
}

type ProjectMirrors []ProjectMirror

type ProjectMirror struct {
	ID                     int         `json:"id"`
	Enabled                bool        `json:"enabled"`
	URL                    string      `json:"url"`
	UpdateStatus           string      `json:"update_status"`
	LastUpdateAt           time.Time   `json:"last_update_at"`
	LastUpdateStartedAt    time.Time   `json:"last_update_started_at"`
	LastSuccessfulUpdateAt time.Time   `json:"last_successful_update_at"`
	LastError              interface{} `json:"last_error"`
	OnlyProtectedBranches  bool        `json:"only_protected_branches"`
	KeepDivergentRefs      interface{} `json:"keep_divergent_refs"`
}
