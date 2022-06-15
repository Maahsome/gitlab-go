package gitlab

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/maahsome/gron"
	"github.com/olekukonko/tablewriter"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type GroupList []Group

type Group struct {
	ID                             int         `json:"id"`
	WebURL                         string      `json:"web_url"`
	Name                           string      `json:"name"`
	Path                           string      `json:"path"`
	Description                    string      `json:"description"`
	Visibility                     string      `json:"visibility"`
	ShareWithGroupLock             bool        `json:"share_with_group_lock"`
	RequireTwoFactorAuthentication bool        `json:"require_two_factor_authentication"`
	TwoFactorGracePeriod           int         `json:"two_factor_grace_period"`
	ProjectCreationLevel           string      `json:"project_creation_level"`
	AutoDevopsEnabled              interface{} `json:"auto_devops_enabled"`
	SubgroupCreationLevel          string      `json:"subgroup_creation_level"`
	EmailsDisabled                 interface{} `json:"emails_disabled"`
	MentionsDisabled               interface{} `json:"mentions_disabled"`
	LfsEnabled                     bool        `json:"lfs_enabled"`
	DefaultBranchProtection        int         `json:"default_branch_protection"`
	AvatarURL                      interface{} `json:"avatar_url"`
	RequestAccessEnabled           bool        `json:"request_access_enabled"`
	FullName                       string      `json:"full_name"`
	FullPath                       string      `json:"full_path"`
	CreatedAt                      time.Time   `json:"created_at"`
	ParentID                       int         `json:"parent_id"`
	LdapCn                         interface{} `json:"ldap_cn"`
	LdapAccess                     interface{} `json:"ldap_access"`
	MarkedForDeletionOn            interface{} `json:"marked_for_deletion_on"`
}

// ToJSON - Write the output as JSON
func (gr *GroupList) ToJSON() string {
	grJSON, err := json.MarshalIndent(gr, "", "  ")
	if err != nil {
		logrus.WithError(err).Error("Error extracting JSON")
		return ""
	}
	return string(grJSON[:])
}

func (gr *GroupList) ToGRON() string {
	grJSON, err := json.MarshalIndent(gr, "", "  ")
	if err != nil {
		logrus.WithError(err).Error("Error extracting JSON for GRON")
	}
	subReader := strings.NewReader(string(grJSON[:]))
	subValues := &bytes.Buffer{}
	ges := gron.NewGron(subReader, subValues)
	ges.SetMonochrome(false)
	if serr := ges.ToGron(); serr != nil {
		logrus.WithError(serr).Error("Problem generating GRON syntax")
		return ""
	}
	return string(subValues.Bytes())
}

func (gr *GroupList) ToYAML() string {
	grYAML, err := yaml.Marshal(gr)
	if err != nil {
		logrus.WithError(err).Error("Error extracting YAML")
		return ""
	}
	return string(grYAML[:])
}

func (gr *GroupList) ToTEXT(noHeaders bool, user string) string {
	buf, row := new(bytes.Buffer), make([]string, 0)

	// ************************** TableWriter ******************************
	table := tablewriter.NewWriter(buf)
	if !noHeaders {
		table.SetHeader([]string{"ID", "GROUP", "BASH"})
		table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	}

	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("\t") // pad with tabs
	table.SetNoWhiteSpace(true)

	// for i=0; i<=len(gr); i++ {
	for _, v := range *gr {
		var cmdLine string
		if len(user) > 0 {
			cmdLine = fmt.Sprintf("<bash:gitlab-tool get project -g %d -u %s>", v.ID, user)
		} else {
			cmdLine = fmt.Sprintf("<bash:gitlab-tool get project -g %d>", v.ID)
		}
		row = []string{
			fmt.Sprintf("%d", v.ID),
			v.FullPath,
			cmdLine,
		}
		table.Append(row)
	}

	table.Render()

	return buf.String()

}

// ToJSON - Write the output as JSON
func (gr *Group) ToJSON() string {
	grJSON, err := json.MarshalIndent(gr, "", "  ")
	if err != nil {
		logrus.WithError(err).Error("Error extracting JSON")
		return ""
	}
	return string(grJSON[:])
}

func (gr *Group) ToGRON() string {
	grJSON, err := json.MarshalIndent(gr, "", "  ")
	if err != nil {
		logrus.WithError(err).Error("Error extracting JSON for GRON")
	}
	subReader := strings.NewReader(string(grJSON[:]))
	subValues := &bytes.Buffer{}
	ges := gron.NewGron(subReader, subValues)
	ges.SetMonochrome(false)
	if serr := ges.ToGron(); serr != nil {
		logrus.WithError(serr).Error("Problem generating GRON syntax")
		return ""
	}
	return string(subValues.Bytes())
}

func (gr *Group) ToYAML() string {
	grYAML, err := yaml.Marshal(gr)
	if err != nil {
		logrus.WithError(err).Error("Error extracting YAML")
		return ""
	}
	return string(grYAML[:])
}

func (gr *Group) ToTEXT(noHeaders bool, user string) string {
	buf, row := new(bytes.Buffer), make([]string, 0)

	// ************************** TableWriter ******************************
	table := tablewriter.NewWriter(buf)
	if !noHeaders {
		table.SetHeader([]string{"ID", "GROUP", "BASH"})
		table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	}

	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("\t") // pad with tabs
	table.SetNoWhiteSpace(true)

	// for _, v := range *gr {
	var cmdLine string
	if len(user) > 0 {
		cmdLine = fmt.Sprintf("<bash:gitlab-tool get project -g %d -u %s>", gr.ID, user)
	} else {
		cmdLine = fmt.Sprintf("<bash:gitlab-tool get project -g %d>", gr.ID)
	}
	row = []string{
		fmt.Sprintf("%d", gr.ID),
		gr.FullPath,
		cmdLine,
	}
	table.Append(row)
	// }

	table.Render()

	return buf.String()

}
