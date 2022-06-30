package utils_test

import (
	"config-manager/utils"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestVerifyStatePayload(t *testing.T) {
	tests := []struct {
		desc  string
		input struct {
			current map[string]string
			payload map[string]string
		}
		want      bool
		wantError error
	}{
		{
			desc: "valid payload",
			input: struct {
				current map[string]string
				payload map[string]string
			}{
				current: map[string]string{
					"insights":            "enabled",
					"remediations":        "enabled",
					"compliance_openscap": "enabled",
				},
				payload: map[string]string{
					"insights":            "enabled",
					"remediations":        "enabled",
					"compliance_openscap": "disabled",
				},
			},
			want: false,
		},
		{
			desc: "payload equal to current state",
			input: struct {
				current map[string]string
				payload map[string]string
			}{
				current: map[string]string{
					"insights":            "enabled",
					"remediations":        "enabled",
					"compliance_openscap": "enabled",
				},
				payload: map[string]string{
					"insights":            "enabled",
					"remediations":        "enabled",
					"compliance_openscap": "enabled",
				},
			},
			want: true,
		},
		{
			desc: "additional services enabled when insights is disabled",
			input: struct {
				current map[string]string
				payload map[string]string
			}{
				current: map[string]string{
					"insights":            "enabled",
					"remediations":        "enabled",
					"compliance_openscap": "enabled",
				},
				payload: map[string]string{
					"insights":            "disabled",
					"remediations":        "enabled",
					"compliance_openscap": "enabled",
				},
			},
			want:      false,
			wantError: cmpopts.AnyError,
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			got, err := utils.VerifyStatePayload(test.input.current, test.input.payload)
			if test.wantError != nil {
				if !cmp.Equal(test.wantError, err, cmpopts.EquateErrors()) {
					t.Errorf("%#v != %#v", test.wantError, err)
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				if test.want != got {
					t.Errorf("%v != %v", test.want, got)
				}
			}
		})
	}
}
