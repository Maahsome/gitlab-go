package gitlab

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/maahsome/gron"
	"github.com/muesli/termenv"
	"github.com/olekukonko/tablewriter"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type Pipelines []Pipeline

type Pipeline struct {
	ID        int       `json:"id"`
	ProjectID int       `json:"project_id"`
	Sha       string    `json:"sha"`
	Ref       string    `json:"ref"`
	Status    string    `json:"status"`
	Source    string    `json:"source"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	WebURL    string    `json:"web_url"`
}

// ToJSON - Write the output as JSON
func (pl *Pipelines) ToJSON() string {
	plJSON, err := json.MarshalIndent(pl, "", "  ")
	if err != nil {
		logrus.WithError(err).Error("Error extracting JSON")
		return ""
	}
	return string(plJSON[:])
}

func (pl *Pipelines) ToGRON() string {
	plJSON, err := json.MarshalIndent(pl, "", "  ")
	if err != nil {
		logrus.WithError(err).Error("Error extracting JSON for GRON")
	}
	subReader := strings.NewReader(string(plJSON[:]))
	subValues := &bytes.Buffer{}
	ges := gron.NewGron(subReader, subValues)
	ges.SetMonochrome(false)
	if serr := ges.ToGron(); serr != nil {
		logrus.WithError(serr).Error("Problem generating GRON syntax")
		return ""
	}
	return string(subValues.Bytes())
}

func (pl *Pipelines) ToYAML() string {
	plYAML, err := yaml.Marshal(pl)
	if err != nil {
		logrus.WithError(err).Error("Error extracting YAML")
		return ""
	}
	return string(plYAML[:])
}

func (pl *Pipelines) ToTEXT(noHeaders bool) string {
	term := termenv.NewOutput(os.Stdout)
	buf, _ := new(bytes.Buffer), make([]string, 0)

	// ************************** TableWriter ******************************
	table := tablewriter.NewWriter(buf)
	if !noHeaders {
		table.SetHeader([]string{"ID", "PROJECT_ID", "STATUS", "JOBS"})
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

	for _, v := range *pl {

		row := []string{
			fmt.Sprintf("%d", v.ID),
			fmt.Sprintf("%d", v.ProjectID),
			v.Status,
			term.Hyperlink(fmt.Sprintf("<bash:gitlab-tool get jobs -p %d -l %d>", v.ProjectID, v.ID), fmt.Sprintf("%d-%d", v.ProjectID, v.ID)),
		}
		table.Append(row)
	}

	table.Render()

	return buf.String()

}
