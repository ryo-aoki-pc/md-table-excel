package main

import (
	"fmt"
	"github.com/ryo-aoki-pc/md-table-excel/internal/cli"
	"os"
	"os/signal"
	"runtime"
)

// cli.main と __main__ に対応する入口。
func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	go func() {
		<-interrupt
		ending := "\n"
		if runtime.GOOS == "windows" {
			ending = "\r\n"
		}
		fmt.Fprint(os.Stderr, "中止しました"+ending)
		os.Exit(1)
	}()
	os.Exit(cli.Main(os.Args[1:]))
}
