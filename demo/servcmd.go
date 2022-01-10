package main

import (
	"context"

	"github.com/kingmore96/IM/demo/server"
	"github.com/spf13/cobra"
)

// ServerStartOptions ServerStartOptions
type ServerStartOptions struct {
	id     string
	listen string
}

// NewServerStartCmd creates a new http server command
func NewServerStartCmd(ctx context.Context, version string) *cobra.Command {
	opts := &ServerStartOptions{}

	cmd := &cobra.Command{
		Use:   "chat",
		Short: "Starts a chat server",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServerStart(ctx, opts, version)
		},
	}
	cmd.PersistentFlags().StringVarP(&opts.id, "serverid", "i", "demo", "server id")
	cmd.PersistentFlags().StringVarP(&opts.listen, "listen", "l", ":8000", "listen address")
	return cmd
}

// RunServerStart run http server
func runServerStart(ctx context.Context, opts *ServerStartOptions, version string) error {
	serv := server.NewServer(opts.id, opts.listen)
	defer serv.Shutdown()
	return serv.Start()
}
