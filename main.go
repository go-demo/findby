package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/LyricTian/findby/find"

	"github.com/codegangsta/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "findby"
	app.Author = "Lyric"
	app.Version = "0.1.0"
	app.Usage = "Concurrency find file."
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "regexp, r",
			Usage: "使用正则表达式过滤文件内容",
		},
		cli.StringSliceFlag{
			Name:  "name, n",
			Usage: "指定过滤的目录名或文件名",
		},
		cli.StringFlag{
			Name:  "out, o",
			Usage: "指定输出文件",
		},
		cli.StringSliceFlag{
			Name:  "ext, e",
			Usage: "指定过滤的文件扩展名",
		},
		cli.IntFlag{
			Name:  "count, c",
			Value: 50,
			Usage: "指定每次并发读取的文件数量",
		},
	}
	app.Action = action
	app.RunAndExitOnError()
}

func tick(ch <-chan time.Time, startTime time.Time) {
	for t := range ch {
		fmt.Print(fmt.Sprintf("\r===> 正在进行文件查找,用时：%.2f s,Goroutine:%d ", float64(t.Sub(startTime))/float64(time.Second), runtime.NumGoroutine()))
	}
}

func action(ctx *cli.Context) {
	reg := ctx.String("regexp")
	names := ctx.StringSlice("name")
	out := ctx.String("out")
	if len(reg) == 0 || len(names) == 0 || len(out) == 0 {
		cli.ShowAppHelp(ctx)
		return
	}
	outPath, err := filepath.Abs(out)
	if err != nil {
		panic(err)
	}
	outFile, err := os.OpenFile(outPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		panic(err)
	}
	defer outFile.Close()

	var (
		startTime = time.Now()
		ticker    = time.NewTicker(time.Millisecond)
	)

	go tick(ticker.C, startTime)
	f := find.NewFile(names, ctx.StringSlice("ext"), reg, ctx.Int("count"))
	for fc := range f.Find() {
		// fmt.Print("\r正在查找文件，用时：", float64(time.Now().Sub(startTime))/float64(time.Millisecond), ",Goroutine:", runtime.NumGoroutine())
		writer := bufio.NewWriter(outFile)
		writer.WriteString(fc.FileName)
		writer.WriteByte('\n')
		writer.WriteByte('\n')
		for _, l := range fc.Lines {
			writer.WriteString(fmt.Sprintf("%d  %s", l.Number, l.Content))
			writer.WriteByte('\n')
		}
		writer.WriteByte('\n')
		writer.Flush()
	}

	fmt.Print(fmt.Sprintf("\r===> 文件查找完成,总用时：%.1f s ", float64(time.Now().Sub(startTime))/float64(time.Second)))
}
