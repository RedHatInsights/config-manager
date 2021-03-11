package utils_test

import (
	"config-manager/domain"
	"config-manager/utils"
	"sort"
	"testing"

	"gotest.tools/assert"
)

func TestInsightsFirst(t *testing.T) {
	state := domain.StateMap{
		"a":        "enabled",
		"b":        "enabled",
		"insights": "enabled",
		"c":        "enabled",
		"z":        "enabled",
	}

	services := state.GetKeys()
	sort.Sort(utils.InsightsFirst(services))
	assert.Equal(t, services[0], "insights")
}
