package cmd

import (
	"github.com/jenkins-x/jx-helpers/v3/pkg/cobras"
	"github.com/jenkins-x/jx-helpers/v3/pkg/options"
	"github.com/jenkins-x/jx-logging/v3/pkg/log"
	"github.com/spf13/cobra"
	"github.com/spring-financial-group/jx-semanticcheck/pkg/cmd/check"
	"github.com/spring-financial-group/jx-semanticcheck/pkg/cmd/version"
	"github.com/spring-financial-group/jx-semanticcheck/pkg/rootcmd"
)

// Options a few common options we tend to use in command line tools
type Options struct {
	options.BaseOptions
}

// Main creates the new command
func Main() *cobra.Command {
	cmd := &cobra.Command{
		Use:   rootcmd.TopLevelCommand,
		Short: "Command for working with Semantic Commits",
		Run: func(cmd *cobra.Command, args []string) {
			err := cmd.Help()
			if err != nil {
				log.Logger().Errorf(err.Error())
			}
		},
	}
	o := options.BaseOptions{}
	o.AddBaseFlags(cmd)

	cmd.AddCommand(cobras.SplitCommand(check.NewCmdCheckSemantics()))
	cmd.AddCommand(cobras.SplitCommand(version.NewCmdVersion()))
	return cmd
}
