package command

import (
	"context"
	"fmt"
	"strings"

	"github.com/urfave/cli/v3"
)

const (
	BoostFlagUsers = "user"
)

type BoostCommandArgs struct {
	Users []string
}

type BoostCommand struct{}

func NewBoostCommand() *BoostCommand {
	return &BoostCommand{}
}

func NewBoostCommandArgs(c *cli.Command) *BoostCommandArgs {
	return &BoostCommandArgs{
		Users: c.StringSlice(BoostFlagUsers),
	}
}

func (cmd *BoostCommand) ToCliCommand() *cli.Command {
	return &cli.Command{
		Name:  "boost",
		Usage: "Boost posts from a list of users",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:     BoostFlagUsers,
				Aliases:  []string{"u"},
				Usage:    "User to boost",
				Required: true,
				Sources:  cli.EnvVars("BOOST_USERS"),
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			args := NewBoostCommandArgs(c)
			fmt.Fprintf(c.Root().Writer, "Boosting posts from %v users: %v\n", len(args.Users), strings.Join(args.Users, ", "))
			return nil
		},
	}
}
