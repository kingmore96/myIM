package main

import (
	"context"
	"fmt"
	"time"

	"github.com/kingmore96/IM/demo/client"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// StartOptions StartOptions
type StartOptions struct {
	address string
	user    string
}

// NewCmd NewCmd
func NewClientCmd(ctx context.Context) *cobra.Command {
	opts := &StartOptions{}

	cmd := &cobra.Command{
		Use:   "client",
		Short: "Start client",
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(ctx, opts)
		},
	}
	cmd.PersistentFlags().StringVarP(&opts.address, "address", "a", "ws://127.0.0.1:8000", "server address")
	cmd.PersistentFlags().StringVarP(&opts.user, "user", "u", "", "user")
	return cmd
}

func run(ctx context.Context, opts *StartOptions) error {
	url := fmt.Sprintf("%s?user=%s", opts.address, opts.user)
	logrus.Info("connect to ", url)

	//连接
	h, err := client.Connect(url)
	if err != nil {
		return err
	}

	//定时发数据，监听close事件
	tk := time.NewTicker(time.Second * 6)
	for {
		select {
		case <-tk.C:
			//发送数据
			err := h.SendText("hello" + "I am" + opts.user)
			if err != nil {
				logrus.Error("sendText - ", err)
			}
		case <-h.Closed:
			logrus.Printf("connection closed")
			return nil
		}
	}
}
