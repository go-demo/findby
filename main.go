package main

import "github.com/codegangsta/cli"

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
			Name:  "ext, e",
			Usage: "指定过滤的文件扩展名",
		},
	}
	app.Action = func(ctx *cli.Context) {
		reg := ctx.String("regexp")
		names := ctx.StringSlice("name")
		if len(reg) == 0 || len(names) == 0 {
			cli.ShowAppHelp(ctx)
			return
		}
	}
	app.RunAndExitOnError()
}
