package audit

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/logrusorgru/aurora"
	"github.com/shopspring/decimal"
	"github.com/sonatype-nexus-community/go-sona-types/ossindex/types"
)

var (
	nine, seven, four decimal.Decimal
)

func init() {
	nine, _ = decimal.NewFromString("9")
	seven, _ = decimal.NewFromString("7")
	four, _ = decimal.NewFromString("4")
}

func LogVulnerablePackage(sb *strings.Builder, noColor bool, idx int, packageCount int, coordinate types.Coordinate) {
	au := aurora.NewAurora(!noColor)
	sb.WriteString(fmt.Sprintf(
		"[%d/%d]\t%s\n%s \n",
		idx,
		packageCount,
		au.Bold(au.Red(coordinate.Coordinates)).String(),
		au.Red(strconv.Itoa(len(coordinate.Vulnerabilities))+" known vulnerabilities affecting installed version").String(),
	))

	sort.Slice(coordinate.Vulnerabilities, func(i, j int) bool {
		return coordinate.Vulnerabilities[i].CvssScore.GreaterThan(coordinate.Vulnerabilities[j].CvssScore)
	})

	for _, v := range coordinate.Vulnerabilities {
		if !v.Excluded {
			t := table.NewWriter()
			t.SetStyle(table.StyleBold)
			t.SetTitle(printColorBasedOnCvssScore(v.CvssScore, v.Title, noColor))
			t.AppendRow([]interface{}{"Description", text.WrapSoft(v.Description, 75)})
			t.AppendSeparator()
			t.AppendRow([]interface{}{"OSS Index ID", v.ID})
			t.AppendSeparator()
			t.AppendRow([]interface{}{"CVSS Score", fmt.Sprintf("%s/10 (%s)", v.CvssScore, scoreAssessment(v.CvssScore))})
			t.AppendSeparator()
			t.AppendRow([]interface{}{"CVSS Vector", v.CvssVector})
			t.AppendSeparator()
			t.AppendRow([]interface{}{"Link for more info", text.WrapHard(v.Reference, 100)})
			sb.WriteString(t.Render() + "\n")
		}
	}
}

func printColorBasedOnCvssScore(score decimal.Decimal, text string, noColor bool) string {
	au := aurora.NewAurora(!noColor)
	if score.GreaterThanOrEqual(nine) {
		return au.Red(au.Bold(text)).String()
	}
	if score.GreaterThanOrEqual(seven) {
		return au.Red(text).String()
	}
	if score.GreaterThanOrEqual(four) {
		return au.Yellow(text).String()
	}
	return au.Green(text).String()
}

func scoreAssessment(score decimal.Decimal) string {
	if score.GreaterThanOrEqual(nine) {
		return "Critical"
	}
	if score.GreaterThanOrEqual(seven) {
		return "High"
	}
	if score.GreaterThanOrEqual(four) {
		return "Medium"
	}
	return "Low"
}
