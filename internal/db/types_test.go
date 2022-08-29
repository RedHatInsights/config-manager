package db

import (
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
)

func TestProfileMarshalJSON(t *testing.T) {
	tests := []struct {
		description string
		input       Profile
		want        []byte
		wantError   error
	}{
		{
			input: Profile{
				ID:           uuid.MustParse("a863569d-6e57-4082-ba80-20e0089738ff"),
				Name:         &JSONNullString{NullString: sql.NullString{Valid: true, String: "default"}},
				Label:        &JSONNullString{NullString: sql.NullString{Valid: true, String: "91fb2f8b"}},
				AccountID:    &JSONNullString{NullString: sql.NullString{Valid: true, String: "123456"}},
				OrgID:        &JSONNullString{NullString: sql.NullString{Valid: true, String: "654321"}},
				CreatedAt:    time.Unix(0, 0).UTC(),
				Active:       true,
				Creator:      &JSONNullString{NullString: sql.NullString{Valid: true, String: "root"}},
				Insights:     true,
				Remediations: true,
				Compliance:   true,
			},
			want: []byte(`{"profile_id":"a863569d-6e57-4082-ba80-20e0089738ff","name":"default","label":"91fb2f8b","account_id":"123456","org_id":"654321","created_at":"1970-01-01T00:00:00Z","active":true,"creator":"root","insights":true,"remediations":true,"compliance":true}`),
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			got, err := json.Marshal(test.input)

			if test.wantError != nil {
				if !cmp.Equal(err, test.wantError, cmpopts.EquateErrors()) {
					t.Errorf("%#v != %#v", err, test.wantError)
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				if !cmp.Equal(got, test.want) {
					t.Errorf("%v", cmp.Diff(got, test.want))
				}
			}
		})
	}
}

func TestProfileUnmarshalJSON(t *testing.T) {
	tests := []struct {
		description string
		input       []byte
		want        Profile
		wantError   error
	}{
		{
			input: []byte(`{"profile_id":"a863569d-6e57-4082-ba80-20e0089738ff","name":"default","label":"91fb2f8b","account_id":"123456","org_id":"654321","created_at":"1970-01-01T00:00:00Z","active":true,"creator":"root","insights":true,"remediations":true,"compliance":true}`),
			want: Profile{
				ID:           uuid.MustParse("a863569d-6e57-4082-ba80-20e0089738ff"),
				Name:         &JSONNullString{NullString: sql.NullString{Valid: true, String: "default"}},
				Label:        &JSONNullString{NullString: sql.NullString{Valid: true, String: "91fb2f8b"}},
				AccountID:    &JSONNullString{NullString: sql.NullString{Valid: true, String: "123456"}},
				OrgID:        &JSONNullString{NullString: sql.NullString{Valid: true, String: "654321"}},
				CreatedAt:    time.Unix(0, 0).UTC(),
				Active:       true,
				Creator:      &JSONNullString{NullString: sql.NullString{Valid: true, String: "root"}},
				Insights:     true,
				Remediations: true,
				Compliance:   true,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			var got Profile
			err := json.Unmarshal(test.input, &got)

			if test.wantError != nil {
				if !cmp.Equal(err, test.wantError, cmpopts.EquateErrors()) {
					t.Errorf("%#v != %#v", err, test.wantError)
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				if !cmp.Equal(got, test.want) {
					t.Errorf("%v", cmp.Diff(got, test.want))
				}
			}
		})
	}
}

func TestJSONNullBoolMarshalJSON(t *testing.T) {
	tests := []struct {
		description string
		input       JSONNullBool
		want        []byte
	}{
		{
			description: "valid = true, bool = true",
			input: JSONNullBool{
				sql.NullBool{
					Valid: true,
					Bool:  true,
				},
			},
			want: []byte(`true`),
		},
		{
			description: "valid = true, bool = false",
			input: JSONNullBool{
				sql.NullBool{
					Valid: true,
					Bool:  false,
				},
			},
			want: []byte(`false`),
		},
		{
			description: "valid = false, bool = true",
			input: JSONNullBool{
				sql.NullBool{
					Valid: false,
					Bool:  true,
				},
			},
			want: []byte(`null`),
		},
		{
			description: "valid = false, bool = false",
			input: JSONNullBool{
				sql.NullBool{
					Valid: false,
					Bool:  false,
				},
			},
			want: []byte(`null`),
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			got, err := test.input.MarshalJSON()
			if err != nil {
				t.Fatal(err)
			}

			if !json.Valid(got) {
				t.Errorf("got invalid JSON: %v", string(got))
			}

			if !cmp.Equal(got, test.want) {
				t.Errorf("%#v != %#v", got, test.want)
			}
		})
	}
}

func TestJSONNullBoolUnmarshalJSON(t *testing.T) {
	tests := []struct {
		description string
		input       []byte
		want        JSONNullBool
	}{
		{
			description: "valid JSON - true",
			input:       []byte(`true`),
			want: JSONNullBool{
				sql.NullBool{
					Valid: true,
					Bool:  true,
				},
			},
		},
		{
			description: "valid JSON - false",
			input:       []byte(`false`),
			want: JSONNullBool{
				NullBool: sql.NullBool{
					Valid: true,
					Bool:  false,
				},
			},
		},
		{
			description: "valid JSON - null",
			input:       []byte(`null`),
			want: JSONNullBool{
				NullBool: sql.NullBool{
					Valid: false,
					Bool:  false,
				},
			},
		},
		{
			description: "invalid JSON - TRUE",
			input:       []byte(`TRUE`),
			want: JSONNullBool{
				NullBool: sql.NullBool{
					Valid: false,
					Bool:  false,
				},
			},
		},
		{
			description: "invalid JSON - ;",
			input:       []byte(`;`),
			want: JSONNullBool{
				NullBool: sql.NullBool{
					Valid: false,
					Bool:  false,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			var got JSONNullBool

			err := got.UnmarshalJSON(test.input)
			if err != nil {
				t.Fatal(err)
			}

			if !cmp.Equal(got, test.want) {
				t.Errorf("%#v != %#v", got, test.want)
			}
		})
	}
}

func TestJSONNullStringMarshalJSON(t *testing.T) {
	tests := []struct {
		description string
		input       JSONNullString
		want        []byte
		wantError   error
	}{
		{
			description: "valid, non-empty string",
			input: JSONNullString{
				NullString: sql.NullString{
					Valid:  true,
					String: `abcd`,
				},
			},
			want: []byte(`"abcd"`),
		},
		{
			description: "valid, empty string",
			input: JSONNullString{
				NullString: sql.NullString{
					Valid:  true,
					String: ``,
				},
			},
			want: []byte(`""`),
		},
		{
			description: "invalid, non-empty string",
			input: JSONNullString{
				NullString: sql.NullString{
					Valid:  false,
					String: `"abcd"`,
				},
			},
			want: []byte(`""`),
		},
		{
			description: "invalid, empty string",
			input: JSONNullString{
				NullString: sql.NullString{
					Valid:  false,
					String: `""`,
				},
			},
			want: []byte(`""`),
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			got, err := test.input.MarshalJSON()

			if test.wantError != nil {
				if !cmp.Equal(err, test.wantError, cmpopts.EquateErrors()) {
					t.Errorf("%#v != %#v", err, test.wantError)
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}

				if !json.Valid(got) {
					t.Errorf("got invalid JSON: %#v", string(got))
				}

				if !cmp.Equal(got, test.want) {
					t.Errorf("%v != %v", got, test.want)
				}
			}
		})
	}
}

func TestJSONNullStringUnmarshalJSON(t *testing.T) {
	tests := []struct {
		description string
		input       []byte
		want        JSONNullString
	}{
		{
			description: "non-empty string input",
			input:       []byte(`"abcd"`),
			want: JSONNullString{
				NullString: sql.NullString{
					Valid:  true,
					String: "abcd",
				},
			},
		},
		{
			description: "empty string input",
			input:       []byte(`""`),
			want: JSONNullString{
				NullString: sql.NullString{
					Valid:  true,
					String: "",
				},
			},
		},
		{
			description: "null input",
			input:       []byte(`null`),
			want: JSONNullString{
				NullString: sql.NullString{
					Valid:  false,
					String: "",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			if !json.Valid(test.input) {
				t.Errorf("got invalid JSON: %#v", string(test.input))
			}

			var got JSONNullString
			err := got.UnmarshalJSON(test.input)
			if err != nil {
				t.Fatal(err)
			}

			if !cmp.Equal(got, test.want) {
				t.Errorf("%#v != %#v", got, test.want)
			}
		})
	}
}
