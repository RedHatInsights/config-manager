package utils_test

import (
	"config-manager/utils"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestInsightsFirst(t *testing.T) {
	tests := []struct {
		desc  string
		input utils.InsightsFirst
		want  utils.InsightsFirst
	}{
		{
			input: utils.InsightsFirst{"a", "b", "insights", "c", "z"},
			want:  utils.InsightsFirst{"insights", "a", "b", "c", "z"},
		},
		{
			input: utils.InsightsFirst{"a", "b", "i", "insights", "c"},
			want:  utils.InsightsFirst{"insights", "a", "b", "i", "c"},
		},
		{
			input: utils.InsightsFirst{"z", "x", "y"},
			want:  utils.InsightsFirst{"z", "x", "y"},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			sort.Sort(test.input)
			if !cmp.Equal(test.input, test.want) {
				t.Errorf("+++%v\n---%v\n%v", test.input, test.want, cmp.Diff(test.input, test.want))
			}
		})
	}
}
