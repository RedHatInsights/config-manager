package util

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestBatchAll(t *testing.T) {
	type Range struct {
		start, end int
	}

	tests := []struct {
		description string
		input       struct {
			count     int
			batchSize int
		}
		want      []Range
		wantError error
	}{
		{
			description: "count equals batch size",
			input: struct {
				count     int
				batchSize int
			}{
				count:     1,
				batchSize: 1,
			},
			want: []Range{{0, 0}},
		},
		{
			description: "count less than batch size",
			input: struct {
				count     int
				batchSize int
			}{
				count:     5,
				batchSize: 10,
			},
			want: []Range{{0, 4}},
		},
		{
			description: "count greater than batch size",
			input: struct {
				count     int
				batchSize int
			}{
				count:     15,
				batchSize: 10,
			},
			want: []Range{{0, 9}, {10, 14}},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			var got []Range
			err := Batch.All(test.input.count, test.input.batchSize, func(start, end int) error {
				got = append(got, Range{start, end})
				return nil
			})

			if test.wantError != nil {
				if !cmp.Equal(err, test.wantError, cmpopts.EquateErrors()) {
					t.Errorf("%#v != %#v", err, test.wantError)
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				if !cmp.Equal(got, test.want, cmp.AllowUnexported(Range{})) {
					t.Errorf("%v", cmp.Diff(got, test.want, cmp.AllowUnexported(Range{})))
				}
			}
		})
	}
}
