package main

import (
	"github.com/google/go-github/github"
	"github.com/jessevdk/go-flags"
	"github.com/uber-go/zap"
	"golang.org/x/oauth2"
	"net/url"
	"os"
)

var opts struct {
	Token     string `short:"t" long:"token" env:"GITHUB_TOKEN" description:"github oauth token" required:"true"`
	Simulate  bool   `short:"s" long:"simulate" description:"don't make any subscription changes"`
	GithubURL string `long:"hub-url" env:"GITHUB_URL" description:"github url (such as your enterprise github url)" default:"https://api.github.com"`
}

var logger zap.Logger

func init() {
	logger = zap.New(zap.NewTextEncoder())

	_, err := flags.Parse(&opts)
	if err != nil {
		logger.Error("unable to parse flags", zap.Error(err))
		os.Exit(1)
	}
}

func main() {
	client := getGithubClient()
	unsubscribeLoop(client)
}

func unsubscribeLoop(client *github.Client) {
	listOpts := github.ListOptions{
		PerPage: 50,
	}
	for {
		repos, resp, err := client.Activity.ListWatched("", &listOpts)
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

func getGithubClient() *github.Client {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: opts.Token})
	tc := oauth2.NewClient(oauth2.NoContext, ts)

	client := github.NewClient(tc)
	baseUrl, err := url.Parse(opts.GithubURL)
	if err != nil {
		logger.Error("unable to parse provided url", zap.String("github url", opts.GithubURL), zap.Error(err))
		os.Exit(1)
	}

	client.BaseURL = baseUrl
	return client
}

func unsubscribeFromRepos(client *github.Client, repos []*github.Repository) {
	for _, repo := range repos {
		ownerStr := repo.Owner.Login
		repoStr := repo.Name
		unsubLogger := logger.With(zap.Bool("simulation", opts.Simulate), zap.String("owner", *ownerStr), zap.String("repo", *repoStr))
		unsubLogger.Info("unsubscribing")
		if !opts.Simulate {
			_, err := client.Activity.DeleteRepositorySubscription(repo.Owner.String(), repo.String())
			if err != nil {
				unsubLogger.Error("error unsubscribing", zap.Error(err))
				os.Exit(1)
			}
		}
	}
}
