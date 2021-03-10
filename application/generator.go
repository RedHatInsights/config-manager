package application

import (
	"bytes"
	"config-manager/domain"
)

// Generator generates a playbook from provided configuration state
type Generator struct {
	PlaybookPath string
	Templates    map[string][]byte
}

func buildKey(name, value string) string {
	switch value {
	case "enabled":
		return name + "_setup.yml"
	case "disabled":
		return name + "_remove.yml"
	default:
		return ""
	}
}

func formatPlay(play []byte) []byte {
	formattedPlay := bytes.Trim(play, "---")
	// formattedPlay = append(formattedPlay, "\n"...)
	return formattedPlay
}

// GeneratePlaybook accepts a state and returns a string representing an ansible playbook.
// TODO: Once an insights playbook exists it should always be first
func (g *Generator) GeneratePlaybook(state domain.StateMap) (string, error) {
	playbook := []byte("---\n# Service Enablement playbook\n")

	if _, exists := state["insights"]; exists {
		insights := g.Templates[buildKey("insights", state["insights"])]
		playbook = append(playbook, formatPlay(insights)...)

		delete(state, "insights") // Add the insights play to the playbook first, then add remaining services
	}

	var err error
	for k, v := range state {
		play := g.Templates[buildKey(k, v)]
		playbook = append(playbook, formatPlay(play)...)
	}

	return string(playbook), err
}
