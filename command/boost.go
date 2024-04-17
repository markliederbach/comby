package command

import (
	"context"
	"fmt"
	"time"

	"github.com/markliederbach/comby/client"
	"github.com/urfave/cli/v3"
)

const (
	BoostFlagUsers          = "user"
	BoostFlagMaxBoosts      = "max-boosts"
	BoostFlagServer         = "server"
	BoostFlagClientID       = "client-id"
	BoostFlagAccessToken    = "access-token"
	BoostFlagSince          = "since"
	BoostFlagExcludePattern = "exclude-pattern"
	BoostFlagExcludeOptions = "exclude"
)

type BoostCommandArgs struct {
	Users           []string
	MastodonServer  string
	AccessToken     string
	MaxBoosts       int
	Since           time.Duration
	ExcludePatterns []string
	ExcludeOptions  []string
}

type BoostCommand struct{}

func NewBoostCommand() *BoostCommand {
	return &BoostCommand{}
}

func NewBoostCommandArgs(c *cli.Command) *BoostCommandArgs {
	return &BoostCommandArgs{
		Users:           c.StringSlice(BoostFlagUsers),
		MastodonServer:  c.String(BoostFlagServer),
		AccessToken:     c.String(BoostFlagAccessToken),
		MaxBoosts:       int(c.Int(BoostFlagMaxBoosts)),
		Since:           c.Duration(BoostFlagSince),
		ExcludePatterns: c.StringSlice(BoostFlagExcludePattern),
		ExcludeOptions:  c.StringSlice(BoostFlagExcludeOptions),
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
			&cli.StringSliceFlag{
				Name:    BoostFlagExcludePattern,
				Usage:   "Exclude statuses matching these patterns",
				Sources: cli.EnvVars("BOOST_EXCLUDE_PATTERNS"),
			},
			&cli.StringSliceFlag{
				Name:  BoostFlagExcludeOptions,
				Usage: "Options to exclude statuses",
				Value: []string{
					"exclude_replies",
					"exclude_boosts",
				},
				DefaultText: "exclude_replies,exclude_boosts",
				Sources:     cli.EnvVars("BOOST_EXCLUDE_OPTIONS"),
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
			filters := GetFilters(mastodonClient, args)
			for _, user := range args.Users {
				account, err := mastodonClient.GetAccount(ctx, user)
				if err != nil {
					return err
				}
				statuses, err := mastodonClient.GetAccountStatuses(ctx, account, filters...)

				if err != nil {
					return err
				}
				statusesToBoost = append(statusesToBoost, statuses...)
			}
			boostedStatuses := 0
			for _, status := range statusesToBoost {
				if boostedStatuses >= args.MaxBoosts {
					fmt.Printf("Reached maximum number of boosts (%d)\n", args.MaxBoosts)
					break
				}
				fmt.Printf("Boosting status %s\n", status.Url)
				boostedStatus, err := mastodonClient.BoostStatus(ctx, status)
				if err != nil {
					return err
				}
				fmt.Printf("Boosted status URL: %s\n", boostedStatus.Url)
				boostedStatuses++
			}
			fmt.Printf("Boosted %d statuses\n", boostedStatuses)
			return nil
		},
	}
}

func GetFilters(mastodonClient *client.MastodonClientImpl, args *BoostCommandArgs) []client.ExcludeFunc {
	filters := []client.ExcludeFunc{
		mastodonClient.ExcludeStatusOlderThan(args.Since),
		mastodonClient.ExcludeAlreadyBoosted(),
		mastodonClient.ExcludeMatchingRegex(args.ExcludePatterns...),
	}

	for _, option := range args.ExcludeOptions {
		switch option {
		case "exclude_replies":
			filters = append(filters, mastodonClient.ExcludeAllReplies())
		case "exclude_replies_to_self":
			filters = append(filters, mastodonClient.ExcludeRepliesToOwnStatuses())
		case "exclude_boosts":
			filters = append(filters, mastodonClient.ExcludeBoosts())
		default:
			fmt.Printf("Unknown exclude option: %s\n", option)
		}
	}
	return filters
}
