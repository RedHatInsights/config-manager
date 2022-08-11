package internal

import (
	"bytes"
	"config-manager/internal/config"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"sync"

	"github.com/rs/zerolog/log"
)

var templates map[string][]byte
var once sync.Once

// GeneratePlaybook returns an Ansible playbook enabling or disabling services
// as defined in the state map.
func GeneratePlaybook(state map[string]string) (string, error) {
	playbook := []byte("---\n# Service Enablement playbook\n")

	keys := make([]string, 0, len(state))
	for k := range state {
		keys = append(keys, k)
	}

	sort.SliceStable(keys, func(i, j int) bool {
		if keys[i] == "insights" {
			return true
		}
		return keys[i] < keys[j]
	})

	var err error
	for _, service := range keys {
		var key string
		switch state[service] {
		case "enabled":
			key = service + "_setup.yml"
		case "disabled":
			key = service + "_remove.yml"
		default:
			return "", fmt.Errorf("cannot generate playbook: unkown service: %v", service)
		}

		once.Do(func() {
			templates = make(map[string][]byte)

			files, err := filepath.Glob(filepath.Join(config.DefaultConfig.PlaybookFiles, "*.yml"))
			if err != nil {
				log.Fatal().Err(err).Msg("cannot match file glob")
			}

			for _, f := range files {
				if fi, err := os.Stat(f); !os.IsNotExist(err) {
					content, err := ioutil.ReadFile(f)
					if err != nil {
						log.Fatal().Err(err).Str("filename", f).Msg("cannot read file")
					}
					templates[fi.Name()] = content
				}
			}
		})

		play, exists := templates[key]
		if exists {
			playbook = append(playbook, bytes.Trim(play, "---")...)
		}
	}

	return string(playbook), err
}

// VerifyStatePayload checks whether currentState and payload are deeply equal,
// additionally qualifying the equality treating the value of the "insights" key
// with higher precedence; if the "insights" key equals "disabled", all keys in
// payload must be "disabled" or an error is returned.
func VerifyStatePayload(currentState, payload map[string]string) (bool, error) {
	equal := false
	if reflect.DeepEqual(currentState, payload) {
		equal = true
		return equal, nil
	}

	if payload["insights"] == "disabled" {
		for k, v := range payload {
			if v != "disabled" {
				return equal, fmt.Errorf("service %s must be disabled if insights is disabled", k)
			}
		}
	}

	return equal, nil
}
