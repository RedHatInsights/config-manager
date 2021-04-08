package application

import (
	"config-manager/domain"
	"config-manager/utils"
	"strings"
	"testing"
)

func TestGenerateAllEnabled(t *testing.T) {
	templates := utils.FilesIntoMap("../playbooks/test/", "*.yml")

	pbg := &Generator{
		Templates: templates,
	}

	state := domain.StateMap{
		"test1": "enabled",
		"test2": "enabled",
	}

	pb, err := pbg.GeneratePlaybook(state)
	if err != nil {
		t.Error(err)
	}

	expectedPlay1 := `
- name: Test 1
  hosts: localhost
  tasks:
    - name: Print 1 setup
      debug:
        msg: "1 setup"`

	expectedPlay2 := `
- name: Test 2
  hosts: localhost
  tasks:
    - name: Print 2 setup
      debug:
        msg: "2 setup"`

	if !strings.Contains(pb, expectedPlay1) || !strings.Contains(pb, expectedPlay2) {
		t.Errorf("Received playbook did not contain expected plays: %s", pb)
	}
}

func TestGenerateAllDisabled(t *testing.T) {
	templates := utils.FilesIntoMap("../playbooks/test/", "*.yml")

	pbg := &Generator{
		Templates: templates,
	}

	state := domain.StateMap{
		"test1": "disabled",
		"test2": "disabled",
	}

	pb, err := pbg.GeneratePlaybook(state)
	if err != nil {
		t.Error(err)
	}

	expectedPlay1 := `
- name: Test 1
  hosts: localhost
  tasks:
    - name: Print 1 removed
      debug:
        msg: "1 removed"`

	expectedPlay2 := `
- name: Test 2
  hosts: localhost
  tasks:
    - name: Print 2 removed
      debug:
        msg: "2 removed"`

	if !strings.Contains(pb, expectedPlay1) || !strings.Contains(pb, expectedPlay2) {
		t.Errorf("Received playbook did not contain expected plays: %s", pb)
	}
}

func TestGenerateWhenServiceDoesNotHavePlay(t *testing.T) {
	templates := utils.FilesIntoMap("../playbooks/test/", "*.yml")

	pbg := &Generator{
		Templates: templates,
	}

	state := domain.StateMap{
		"test1":              "disabled",
		"test2":              "disabled",
		"serviceWithoutPlay": "disabled",
	}

	pb, err := pbg.GeneratePlaybook(state)
	if err != nil {
		t.Error(err)
	}

	expectedPlay1 := `
- name: Test 1
  hosts: localhost
  tasks:
    - name: Print 1 removed
      debug:
        msg: "1 removed"`

	expectedPlay2 := `
- name: Test 2
  hosts: localhost
  tasks:
    - name: Print 2 removed
      debug:
        msg: "2 removed"`

	if !strings.Contains(pb, expectedPlay1) || !strings.Contains(pb, expectedPlay2) {
		t.Errorf("Received playbook did not contain expected plays: %s", pb)
	}
}
