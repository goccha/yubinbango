package cmd

import (
	"context"
	"fmt"
	"github.com/goccha/logging/log"
	"github.com/goccha/logging/masking"
	"github.com/goccha/yubinbango/pkg/env"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use: "app",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func Commands() []*cobra.Command {
	return rootCmd.Commands()
}

func logRequest(ctx context.Context, event interface{}, body string) (err error) {
	if env.DebugLog() {
		var msg []byte
		if msg, err = masking.New("body").Run(ctx, masking.Json(event)); err != nil {
			log.Debug(ctx).Err(err).Send()
		} else {
			log.Debug(ctx).RawJSON("request", msg).Send()
		}
		if msg, err = masking.New("password").Run(ctx, masking.Json(body)); err != nil {
			log.Debug(ctx).Err(err).Send()
		} else if len(msg) > 0 {
			log.Debug(ctx).RawJSON("request", msg).Send()
		}
	}
	return nil
}

var Version = func() *cobra.Command {
	return &cobra.Command{
		Use: "version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(env.Version())
		},
	}
}()

func init() {
	rootCmd.AddCommand(Version)
}
