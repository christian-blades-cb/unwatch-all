package main

import (
	"github.com/google/go-github/github"
	"github.com/jessevdk/go-flags"
	"github.com/uber-go/zap"
	"golang.org/x/oauth2"
	"os"
)

var opts struct {
	Token    string `short:"t" long:"token" env:"GITHUB_TOKEN" description:"github oauth token" required:"true"`
	Simulate bool   `short:"s" long:"simulate" description:"don't make any subscription changes"`
}

var logger zap.Logger

func main() {
	logger = zap.New(zap.NewTextEncoder())

	_, err := flags.Parse(&opts)
	if err != nil {
		logger.Error("unable to parse flags", zap.Error(err))
		os.Exit(1)
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: opts.Token})
	tc := oauth2.NewClient(oauth2.NoContext, ts)

	client := github.NewClient(tc)

	listOpts := github.ListOptions{
		PerPage: 50,
	}
	for {
		repos, resp, err := client.Activity.ListWatched("", nil)
		if err != nil {
			logger.Error("exception while listing watch repos", zap.Error(err))
			os.Exit(1)
		}
		unsubscribeFromRepos(client, repos)

		if resp.NextPage == 0 {
			break
		}
		listOpts.Page = resp.NextPage
	}

}

func unsubscribeFromRepos(client *github.Client, repos []*github.Repository) {
	subscription := &github.Subscription{
		Subscribed: flase(),
	}

	for _, repo := range repos {
		logger.Info("unsubscribing", zap.Bool("simulation", opts.Simulate), zap.String("owner", repo.Owner.String()), zap.String("repo", *repo.Name))
		if !opts.Simulate {
			client.Activity.SetRepositorySubscription(repo.Owner.String(), repo.String(), subscription)
		}
	}
}

func flase() *bool {
	x := false
	return &x
}

func tru() *bool {
	x := true
	return &x
}
