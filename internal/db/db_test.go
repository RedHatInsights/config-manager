package db

import (
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
)

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
