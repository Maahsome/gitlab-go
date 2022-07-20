package gitlab

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/maahsome/gron"
	"github.com/olekukonko/tablewriter"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type Variables []Variable

type Variable struct {
	EnvironmentScope string `json:"environment_scope"`
	Key              string `json:"key"`
	Masked           bool   `json:"masked"`
	Protected        bool   `json:"protected"`
	Value            string `json:"value"`
	VariableType     string `json:"variable_type"`
	Source           string `json:"source"`
}

// ToJSON - Write the output as JSON
func (v *Variables) ToJSON() string {
	vJSON, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		logrus.WithError(err).Error("Error extracting JSON")
		return ""
	}
	return string(vJSON[:])
}

func (v *Variables) ToGRON() string {
	vJSON, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		logrus.WithError(err).Error("Error extracting JSON for GRON")
	}
	subReader := strings.NewReader(string(vJSON[:]))
	subValues := &bytes.Buffer{}
	ges := gron.NewGron(subReader, subValues)
	ges.SetMonochrome(false)
	if serr := ges.ToGron(); serr != nil {
		logrus.WithError(serr).Error("Problem generating GRON syntax")
		return ""
	}
	return string(subValues.Bytes())
}

func (v *Variables) ToYAML() string {
	vYAML, err := yaml.Marshal(v)
	if err != nil {
		logrus.WithError(err).Error("Error extracting YAML")
		return ""
	}
	return string(vYAML[:])
}

func (v *Variables) ToTEXT(noHeaders bool) string {
	buf, _ := new(bytes.Buffer), make([]string, 0)

	// ************************** TableWriter ******************************
	table := tablewriter.NewWriter(buf)
	if !noHeaders {
		table.SetHeader([]string{"VARIABLE", "VALUE"})
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

	for _, i := range *v {

		row := []string{
			i.Key,
			i.Value,
		}
		table.Append(row)
	}

	table.Render()

	return buf.String()

}
