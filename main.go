package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/markliederbach/comby/command"
	"github.com/urfave/cli/v3"
)

var (
	// Version is automatically overridden at compile time
	Version = "latest"
	AppName = "comby"
)

func init() {
	cli.VersionPrinter = func(cmd *cli.Command) {
		fmt.Printf("%s %s\n", cmd.Root().Name, cmd.Root().Version)
	}
}

func main() {
	cmd := cli.Command{
		Name:    "Comby",
		Version: Version,
		Authors: []any{
			"Mark Liederbach",
		},
		Usage: "Automatically boost Mastodon posts from a list of users",
		Commands: []*cli.Command{
			command.NewBoostCommand().ToCliCommand(),
		},
	}
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
