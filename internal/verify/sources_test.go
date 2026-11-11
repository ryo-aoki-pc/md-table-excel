package verify_test

import (
	"strings"
	"testing"

	"github.com/ryo-aoki-pc/md-table-excel/internal/model"
	"github.com/ryo-aoki-pc/md-table-excel/internal/table"
	"github.com/ryo-aoki-pc/md-table-excel/internal/verify"
)

func TestSourceExpectationsRejectByteChanges(t *testing.T) {
	input := "<table class='t'><thead class='h'><tr><th style='color:red'>a</th></tr></thead><tbody><tr><td>b</td></tr></tbody></table>"
	for _, tc := range []struct{ name, text, problem string }{
		{"セルの属性", strings.Replace(input, " style='color:red'", "", 1), "セル（1 行 1 列）の原文"},
		{"行の属性", strings.Replace(input, "<tr>", "<tr class=x>", 1), "<tr> の開始タグ"},
		{"区画", "<table class='t'><thead class='h'><tr><th style='color:red'>a</th></tr><tr><td>b</td></tr></thead><tbody></tbody></table>", "区画"},
		{"表の属性", strings.Replace(input, " class='t'", "", 1), "表の開始タグ"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			md := table.NewMarkdown("", input, false)
			r := table.FindTables(md)[0]
			e := &table.Expectation{Cells: map[model.Position]string{{Row: 0, Col: 0}: "<th style='color:red'>a</th>", {Row: 1, Col: 0}: "<td>b</td>"}, RowTags: []string{"<tr>", "<tr>"}, Sections: []int{1, 1}}
			repl := &table.Replacement{Region: r, Lines: []string{tc.text}, Model: r.Model(), TargetKind: "html", Expect: e}
			out := verify.Check(md, tc.text, []*table.Replacement{repl}, verify.Options{KeepHoles: true})
			if len(out.Problems) == 0 || !strings.Contains(strings.Join(out.Problems, "/"), tc.problem) {
				t.Fatalf("原文の違いを見逃した: %v", out.Problems)
			}
		})
	}
}

func TestSourceExpectationsCheckHolesAndPipeBodies(t *testing.T) {
	input := "<table><tr><td>a</td><td>b</td></tr><tr><td>c</td></tr></table>"
	md := table.NewMarkdown("", input, false)
	r := table.FindTables(md)[0]
	expected := verify.CloneModel(r.Model(), "html")
	e := &table.Expectation{Cells: map[model.Position]string{{Row: 0, Col: 0}: "<td>a</td>", {Row: 0, Col: 1}: "<td>b</td>", {Row: 1, Col: 0}: "<td>c</td>"}, RowTags: []string{"<tr>", "<tr>"}, Sections: []int{2}}
	repl := &table.Replacement{Region: r, Lines: []string{input}, Model: expected, TargetKind: "html", Expect: e}
	out := verify.Check(md, input, []*table.Replacement{repl}, verify.Options{KeepHoles: true})
	if !strings.Contains(strings.Join(out.Problems, "/"), "セルの無い位置") {
		t.Fatalf("穴の違いを見逃した: %v", out.Problems)
	}
	input = "| a |\n| :----: |\n| b | 余分 |"
	md = table.NewMarkdown("", input, false)
	r = table.FindTables(md)[0]
	for _, tc := range []struct{ text, problem string }{{"| a  |\n| :----: |\n| b | 余分 |", "1 行目の原文"}, {"| a |\n| :---: |\n| b | 余分 |", "区切り行の原文"}, {"| a |\n| :----: |\n| b |", "2 行目の原文"}} {
		delimiter := "| :----: |"
		e := &table.Expectation{Cells: map[model.Position]string{{Row: 0, Col: 0}: "a", {Row: 1, Col: 0}: "b"}, Bodies: map[int]string{0: "| a |", 1: "| b | 余分 |"}, Delimiter: &delimiter}
		repl := &table.Replacement{Region: r, Lines: strings.Split(tc.text, "\n"), Model: r.Model(), TargetKind: "pipe", Expect: e}
		out := verify.Check(md, tc.text, []*table.Replacement{repl}, verify.Options{KeepHoles: true})
		if !strings.Contains(strings.Join(out.Problems, "/"), tc.problem) {
			t.Fatalf("パイプの原文の違いを見逃した: %v", out.Problems)
		}
	}
}
