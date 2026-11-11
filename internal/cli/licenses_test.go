package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"

	mdtable "github.com/ryo-aoki-pc/md-table-excel"
)

func TestLicensesAction(t *testing.T) {
	code, out, errout := captureOutput(t, func() int {
		return mainContext(context.Background(), []string{"--licenses"}, func(string) error {
			t.Fatal("Excel を準備しました")
			return nil
		})
	})
	var want bytes.Buffer
	if err := mdtable.WriteLicenses(&want); err != nil {
		t.Fatal(err)
	}
	if code != 0 || errout != "" || out != want.String() {
		t.Fatalf("ライセンスの表示が正しくありません: 終了 %d、標準エラー %q", code, errout)
	}
	if !strings.Contains(commandHelp[""], "--licenses") {
		t.Fatal("使い方に --licenses がありません")
	}
	for _, args := range [][]string{{"--licenses", "a.md"}, {"--help", "--licenses"}, {"--licenses", "--version"}, {"list", "--licenses"}} {
		code, out, errout := captureOutput(t, func() int { return mainContext(context.Background(), args, nil) })
		if code != 2 || out != "" || !strings.Contains(errout, "単独で指定してください") {
			t.Error(args, code, out, errout)
		}
	}
}

func TestLicensesOutputFailure(t *testing.T) {
	code, _, errout := captureOutput(t, func() int {
		if err := stdout.file.Close(); err != nil {
			t.Fatal(err)
		}
		return mainContext(context.Background(), []string{"--licenses"}, nil)
	})
	if code != 1 || errout != "エラー: ライセンスを書き込めません\n" {
		t.Fatal(code, errout)
	}
}
