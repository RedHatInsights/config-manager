package application

import (
	"bytes"
	"config-manager/domain"
	"config-manager/utils"
	"errors"
	"sort"
)

// Generator generates a playbook from provided configuration state
type Generator struct {
	Templates map[string][]byte
}

func buildKey(name, value string) (string, error) {
	switch value {
	case "enabled":
		return name + "_setup.yml", nil
	case "disabled":
		return name + "_remove.yml", nil
	default:
		return "", errors.New("Unknown value: " + value)
	}
}

func formatPlay(play []byte) []byte {
	formattedPlay := bytes.Trim(play, "---")
	return formattedPlay
}

// GeneratePlaybook accepts a state and returns a string representing an ansible playbook.
func (g *Generator) GeneratePlaybook(state domain.StateMap) (string, error) {
	playbook := []byte("---\n# Service Enablement playbook\n")
	services := state.GetKeys()
	sort.Sort(utils.InsightsFirst(services))

	var err error
	for _, service := range services {
		key, err := buildKey(service, state[service])
		if err != nil {
			return "", err
		}

		play, exists := g.Templates[key]
		if exists {
			playbook = append(playbook, formatPlay(play)...)
		}
	}

	return string(playbook), err
}
