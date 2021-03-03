package application

import (
	"bytes"
	"config-manager/domain"
	"io/ioutil"
)

// Generator generates a playbook from provided configuration state
type Generator struct {
	PlaybookPath string
}

func buildFilename(path, name, value string) string {
	if value == "enabled" {
		return path + name + "_setup.yml"
	}

	return path + name + "_remove.yml"
}

func formatPlay(play []byte) []byte {
	formattedPlay := bytes.Trim(play, "---")
	formattedPlay = append(formattedPlay, "\n"...)
	return formattedPlay
}

// GeneratePlaybook accepts a state and returns a string representing an ansible playbook.
// TODO: Once an insights playbook exists it should always be first
func (g *Generator) GeneratePlaybook(state domain.StateMap) (string, error) {
	pb := []byte("---\n# Service Enablement playbook\n") // Read the insights playbook and remove "insights" from the state map before proceeding
	var err error
	for k, v := range state {
		filename := buildFilename(g.PlaybookPath, k, v)
		play, err := ioutil.ReadFile(filename)
		if err != nil {
			return "", err
		}
		play = formatPlay(play)
		pb = append(pb, play...)
	}

	return string(pb), err
}
