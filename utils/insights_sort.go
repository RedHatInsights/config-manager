package utils

// InsightsFirst is a string slice that, when sorted, puts "insights" first if
// found.
type InsightsFirst []string

func (a InsightsFirst) Len() int      { return len(a) }
func (a InsightsFirst) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a InsightsFirst) Less(i, j int) bool {
	switch a[i] {
	case "insights":
		return true
	}
	return false
}
