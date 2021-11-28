package synctest

import (
	"fmt"
	"sort"
)

// NilChecker TODO.
var NilChecker = nilChecker{}

type nilChecker struct{}

func (nc nilChecker) Info() string                  { return "expected == actual" }
func (nc nilChecker) Data() interface{}             { return nil }
func (nc nilChecker) Check(actual interface{}) bool { return actual == nil }

// Checker TODO.
type Checker interface {
	Check(interface{}) bool
	Data() interface{}
	Info() string
}

// SetDiff TODO.
type SetDiff struct {
	Missing     []Checker
	Extra, Same []interface{}
}

// NewSetDiff TODO.
func NewSetDiff(actual []interface{}, expected []Checker) SetDiff {
	res := SetDiff{
		Missing: expected,
		Extra:   make([]interface{}, 0, len(actual)),
		Same:    make([]interface{}, 0, len(actual)),
	}

L:
	for _, actualItem := range actual {
		for i, expectedItem := range res.Missing {
			if expectedItem.Check(actualItem) {
				// Add to same list
				res.Same = append(res.Same, actualItem)

				// Delete from expected list
				lastIdx := len(res.Missing) - 1
				res.Missing[i] = res.Missing[lastIdx]
				res.Missing = res.Missing[:lastIdx]

				// Continue with next actual
				continue L
			}
		}

		// Add to extra list
		res.Extra = append(res.Extra, actualItem)
	}

	return res
}

// AllSame TODO.
func (sd SetDiff) AllSame() bool { return len(sd.Missing) == 0 && len(sd.Extra) == 0 }

func (sd SetDiff) String() string {
	var diff complexDiff

	if len(sd.Missing) > 0 {
		missingMap := make(map[string][]interface{})
		for _, item := range sd.Missing {
			info := item.Info()
			missingMap[info] = append(missingMap[info], item.Data())
		}

		for info, items := range missingMap {
			diff = append(diff, complexDiffSection{
				title:      fmt.Sprintf("expected missing | %s", info),
				itemPrefix: "-",
				items:      items,
			})
		}

		sort.Sort(diff)
	}

	if len(sd.Extra) > 0 {
		diff = append(diff, complexDiffSection{
			title:      "actual extra",
			itemPrefix: "+",
			items:      sd.Extra,
		})
	}

	if len(sd.Same) > 0 {
		diff = append(diff, complexDiffSection{
			title:      "same",
			itemPrefix: "•",
			items:      sd.Same,
		})
	}

	return diff.String()
}
