package utils_test

import (
	"config-manager/utils"
	"testing"

	"github.com/stretchr/testify/assert"
)

var verifyPayloadTests = []struct {
	name         string
	currentState map[string]string
	payload      map[string]string
	equal        bool
	errorThrown  bool
}{
	{
		"valid payload",
		map[string]string{
			"insights":            "enabled",
			"remediations":        "enabled",
			"compliance_openscap": "enabled",
		},
		map[string]string{
			"insights":            "enabled",
			"remediations":        "enabled",
			"compliance_openscap": "disabled",
		},
		false,
		false,
	},
	{
		"payload equal to current state",
		map[string]string{
			"insights":            "enabled",
			"remediations":        "enabled",
			"compliance_openscap": "enabled",
		},
		map[string]string{
			"insights":            "enabled",
			"remediations":        "enabled",
			"compliance_openscap": "enabled",
		},
		true,
		false,
	},
	{
		"additional services enabled when insights is disabled",
		map[string]string{
			"insights":            "enabled",
			"remediations":        "enabled",
			"compliance_openscap": "enabled",
		},
		map[string]string{
			"insights":            "disabled",
			"remediations":        "enabled",
			"compliance_openscap": "enabled",
		},
		false,
		true,
	},
}

func TestVerifyStatePayload(t *testing.T) {
	for _, tt := range verifyPayloadTests {
		t.Run(tt.name, func(t *testing.T) {
			equal, err := utils.VerifyStatePayload(tt.currentState, tt.payload)
			if tt.errorThrown {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}

			if tt.equal {
				assert.True(t, equal)
			} else {
				assert.False(t, equal)
			}
		})
	}
}
