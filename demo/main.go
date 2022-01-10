package main

import (
	"context"
	"flag"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const version = "v1"

func main() {
	flag.Parse()

	root := &cobra.Command{
		Use:     "chat",
		Version: version,
		Short:   "chat demo",
	}
	ctx := context.Background()

	root.AddCommand(NewServerStartCmd(ctx, version))
	root.AddCommand(NewClientCmd(ctx))

	if err := root.Execute(); err != nil {
		logrus.WithError(err).Fatal("Could not run command")
	}
}
