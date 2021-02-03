package synctest

import (
	"fmt"
	"strings"
)

const fmtReport = `%s
> %s
───────────
%s
`

type report struct{ name, info, diff string }

func (r report) String() string { return fmt.Sprintf(fmtReport, r.name, r.info, r.diff) }

const fmtSimpleDiff = `expected: %v
actual:   %v`

type simpleDiff struct{ expected, actual interface{} }

func (sd simpleDiff) String() string { return fmt.Sprintf(fmtSimpleDiff, sd.expected, sd.actual) }

type complexDiffSection struct {
	title, itemPrefix string
	items             []interface{}
}

func (cds complexDiffSection) String() string {
	var b strings.Builder
	b.WriteString(cds.title)
	for _, item := range cds.items {
		fmt.Fprintf(&b, "\n%s %v", cds.itemPrefix, item)
	}
	return b.String()
}

type complexDiff []complexDiffSection

func (cd complexDiff) String() string {
	lines := make([]string, len(cd))
	for i, section := range cd {
		lines[i] = section.String()
	}
	return strings.Join(lines, "\n")
}

func (cd complexDiff) Len() int           { return len(cd) }
func (cd complexDiff) Less(i, j int) bool { return cd[i].title < cd[j].title }
func (cd complexDiff) Swap(i, j int)      { cd[i], cd[j] = cd[j], cd[i] }
