package command

import (
	"context"
	"fmt"
	"time"

	"github.com/markliederbach/comby/client"
	"github.com/urfave/cli/v3"
)

const (
	BoostFlagUsers       = "user"
	BoostFlagMaxBoosts   = "max-boosts"
	BoostFlagServer      = "server"
	BoostFlagClientID    = "client-id"
	BoostFlagAccessToken = "access-token"
	BoostFlagSince       = "since"
)

type BoostCommandArgs struct {
	Users          []string
	MastodonServer string
	AccessToken    string
	MaxBoosts      int64
	Since          time.Duration
}

type BoostCommand struct{}

func NewBoostCommand() *BoostCommand {
	return &BoostCommand{}
}

func NewBoostCommandArgs(c *cli.Command) *BoostCommandArgs {
	return &BoostCommandArgs{
		Users:          c.StringSlice(BoostFlagUsers),
		MastodonServer: c.String(BoostFlagServer),
		AccessToken:    c.String(BoostFlagAccessToken),
		MaxBoosts:      c.Int(BoostFlagMaxBoosts),
		Since:          c.Duration(BoostFlagSince),
	}
}

func (cmd *BoostCommand) ToCliCommand() *cli.Command {
	return &cli.Command{
		Name:  "boost",
		Usage: "Boost posts from a list of users",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     BoostFlagServer,
				Aliases:  []string{"s"},
				Usage:    "Mastodon server to authenticate with",
				Required: true,
				Sources:  cli.EnvVars("MASTODON_SERVER"),
			},
			&cli.StringFlag{
				Name:     BoostFlagAccessToken,
				Aliases:  []string{"t"},
				Usage:    "Mastodon access token",
				Required: true,
				Sources:  cli.EnvVars("MASTODON_ACCESS_TOKEN"),
			},
			&cli.StringSliceFlag{
				Name:     BoostFlagUsers,
				Aliases:  []string{"u"},
				Usage:    "User to boost",
				Required: true,
				Sources:  cli.EnvVars("BOOST_USERS"),
			},
			&cli.IntFlag{
				Name:        BoostFlagMaxBoosts,
				Aliases:     []string{"m"},
				Usage:       "Maximum number of boosts to perform",
				DefaultText: "2",
				Value:       2,
				Sources:     cli.EnvVars("BOOST_MAX_BOOSTS"),
			},
			&cli.DurationFlag{
				Name:        BoostFlagSince,
				Usage:       "Only boost statuses in the last N duration",
				Value:       time.Hour * 24,
				DefaultText: "24h",
				Sources:     cli.EnvVars("BOOST_SINCE"),
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			args := NewBoostCommandArgs(c)
			mastodonClient, err := client.NewMastodonClient(ctx, &client.MastodonClientOptions{
				Server:      args.MastodonServer,
				AccessToken: args.AccessToken,
			})
			if err != nil {
				return err
			}
			statusesToBoost := []client.Status{}

			filters := []client.ExcludeFunc{
				// mastodonClient.ExcludeRepliesToOwnStatuses(),
				mastodonClient.ExcludeStatusOlderThan(args.Since),
			}

			for _, user := range args.Users {
				account, err := mastodonClient.GetAccount(ctx, user)
				if err != nil {
					return err
				}
				statuses, err := mastodonClient.GetAccountStatuses(ctx, account, filters...)

				if err != nil {
					return err
				}
				// TODO: filter statuses
				// 		- exclude replies to own statuses
				// 		- exclude reblogs
				statusesToBoost = append(statusesToBoost, statuses...)
			}
			for _, status := range statusesToBoost {
				fmt.Printf("Boosting status %s\n", status.Id)
			}
			return nil
		},
	}
}
