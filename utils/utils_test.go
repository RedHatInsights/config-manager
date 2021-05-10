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
		true,
	},
}

func TestVerifyStatePayload(t *testing.T) {
	for _, tt := range verifyPayloadTests {
		t.Run(tt.name, func(t *testing.T) {
			err := utils.VerifyStatePayload(tt.currentState, tt.payload)
			if tt.errorThrown {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
