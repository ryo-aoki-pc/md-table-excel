package verify

import (
	"fmt"
	"reflect"

	"github.com/ryo-aoki-pc/md-table-excel/internal/table"
)

// checkSources はモデルが持たない原文の属性・区画・穴を検査する。
func checkSources(r, again *table.Region, repl *table.Replacement) []string {
	e := repl.Expect
	got := table.SourceExpectation(again)
	problems := []string{}
	if r.Kind != again.Kind {
		return []string{fmt.Sprintf("表 %d: 原文の検査で表の種類が変わる", r.Number)}
	}
	if !reflect.DeepEqual(e.Cells, got.Cells) {
		for p, wanted := range e.Cells {
			if raw, ok := got.Cells[p]; !ok || raw != wanted {
				problems = append(problems, fmt.Sprintf("表 %d: セル（%d 行 %d 列）の原文が期待したバイトと違う", r.Number, p.Row+1, p.Col+1))
			}
		}
		if len(e.Cells) != len(got.Cells) {
			problems = append(problems, fmt.Sprintf("表 %d: 原文を検査するセルの個数が変わる", r.Number))
		}
	}
	if r.Kind == "pipe" {
		for row, body := range e.Bodies {
			if got.Bodies[row] != body {
				problems = append(problems, fmt.Sprintf("表 %d: %d 行目の原文が変わる", r.Number, row+1))
			}
		}
		if e.Delimiter != nil && *e.Delimiter != *got.Delimiter {
			problems = append(problems, fmt.Sprintf("表 %d: 区切り行の原文が変わる", r.Number))
		}
		return problems
	}
	if e.RowTags != nil && !reflect.DeepEqual(e.RowTags, got.RowTags) {
		problems = append(problems, fmt.Sprintf("表 %d: 行の <tr> の開始タグが変わる", r.Number))
	}
	old, new := r.HTML, again.HTML
	sectionsSame := len(old.Sections) == len(new.Sections)
	if sectionsSame {
		for i, s := range old.Sections {
			n := new.Sections[i]
			if s.Tag != n.Tag || (s.Token == nil) != (n.Token == nil) || s.Token != nil && s.Token.Text != n.Token.Text {
				sectionsSame = false
				break
			}
		}
	}
	if !sectionsSame || e.Sections != nil && !reflect.DeepEqual(e.Sections, got.Sections) {
		problems = append(problems, fmt.Sprintf("表 %d: 足した行が別の区画に入るか、区画の原文が変わる", r.Number))
	}
	if old.TableToken.Text != new.TableToken.Text || !reflect.DeepEqual(old.Caption, new.Caption) || !reflect.DeepEqual(old.Colgroups, new.Colgroups) || r.Head != again.Head || r.Tail != again.Tail {
		problems = append(problems, fmt.Sprintf("表 %d: 表の開始タグ・説明・列の指定・前後の原文が変わる", r.Number))
	}
	if repl.Model != nil && !reflect.DeepEqual(repl.Model.Holes, again.Model().Holes) {
		problems = append(problems, fmt.Sprintf("表 %d: セルの無い位置が期待した形と違う", r.Number))
	}
	return problems
}
