package main

import (
	"os"

	"github.com/micastar/file-to-storage-and-share/pkg/action"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "ftageshare",
		Usage: "A simple stage and share file CLI",
		Commands: []*cli.Command{
			{
				Name:    "web",
				Aliases: []string{"i"},
				Usage:   "Launch a server",
				Action:  action.LaunchServer,
			},
			{
				Name:   "upload",
				Usage:  "Upload a file",
				Action: action.UploadFile,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "file",
						Aliases:  []string{"f"},
						Usage:    "File path",
						Required: true,
					},
				},
			},
			{
				Name:   "download",
				Usage:  "Download a file",
				Action: action.DownloadFile,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "id",
						Usage:    "File ID",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "output",
						Aliases:  []string{"o"},
						Usage:    "Output path",
						Required: false,
					},
				},
			},
		},
	}

	app.Run(os.Args)
}
