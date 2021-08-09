package main

import (
	"os"

	"github.com/6de1ay/auto-semver-tag/pkg/git"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use: "auto-semver-tag",
	}
	rootCmd.SetOut(os.Stdout)

	rootCmd.AddCommand(command())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func command() *cobra.Command {
	return &cobra.Command{
		Use:  "exec [REPOSITORY] [RELEASE_BRANCH] [COMMIT_SHA] [GH_EVENT_PATH]",
		Args: cobra.ExactArgs(4),
		Run:  executeCommand,
	}
}

func executeCommand(cmd *cobra.Command, args []string) {
	reposiroy := args[0]
	releaseBranch := args[1]
	commitSha := args[2]
	githubEventFilePath := args[3]

	token, isExists := os.LookupEnv("GTIHUB_TOKEN")
	if !isExists {
		panic("token does not exists")
	}

	client := git.New(token, reposiroy, releaseBranch)
	client.PerformAction(commitSha, githubEventFilePath)
}
