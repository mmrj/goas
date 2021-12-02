package main

import (
	"log"
	"os"

	"github.com/urfave/cli"
)

var version = "v1.0.0"

var flags = []cli.Flag{
	cli.StringFlag{
		Name:  "module-path",
		Value: ".",
		Usage: "goas will search @comment under the module",
	},
	cli.StringFlag{
		Name:  "main-file-path",
		Value: "",
		Usage: "goas will start to search @comment from this main file",
	},
	cli.StringFlag{
		Name:  "handler-path",
		Value: "",
		Usage: "goas only search handleFunc comments under the path",
	},
	cli.StringFlag{
		Name:  "file-ref-path",
		Value: "",
		Usage: "path to start looking for file refs",
	},
	cli.StringFlag{
		Name:  "output",
		Value: "oas.json",
		Usage: "output file",
	},
	cli.BoolFlag{
		Name:  "debug",
		Usage: "show debug message",
	},
	cli.BoolFlag{
		Name:  "omit-packages",
		Usage: "Omit packages from schema names. An error will be thrown if there is a conflict.",
	},
	cli.BoolFlag{
		Name:  "show-hidden",
		Usage: "Generate schema even for paths that are marked as hidden packages",
	},
}

func action(c *cli.Context) error {
	p, err := newParser(c.GlobalString("module-path"), c.GlobalString("main-file-path"), c.GlobalString("handler-path"), c.GlobalString("file-ref-path"), c.GlobalBool("debug"), c.GlobalBool("omit-packages"), c.GlobalBool("show-hidden"))
	if err != nil {
		return err
	}

	return p.CreateOASFile(c.GlobalString("output"))
}

func main() {
	app := cli.NewApp()
	app.Name = "goas"
	app.Usage = ""
	// app.UsageText = "goas [options]"
	app.Version = version
	app.Copyright = "(c) 2018 mikun800527@gmail.com"
	app.HideHelp = true
	app.OnUsageError = func(c *cli.Context, err error, isSubcommand bool) error {
		cli.ShowAppHelp(c)
		return nil
	}
	app.Flags = flags
	app.Action = action

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal("Error: ", err)
	}
}
