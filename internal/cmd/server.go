package cmd

import (
	"context"
	"errors"
	"fmt"
	"github.com/goccha/yubinbango/internal/routes"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/goccha/envar"
	ginlog "github.com/goccha/logging/gin"
	"github.com/goccha/logging/log"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(Server)
}

var Server = NewServer()

// NewServer HTTPサーバー起動
func NewServer() *cobra.Command {
	type Options struct {
		DirPath          string
		HealthCheck      bool
		BasicAuth        string
		BasicAuthEnabled bool
	}
	opts := &Options{}
	cmd := &cobra.Command{
		Use:     "server",
		Aliases: []string{"api"},
		Short:   "APIサーバー起動",
		Long:    "API サーバーを起動します",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			router := gin.New()
			router.Use(ginlog.AccessLog(), gin.Recovery())
			port := envar.Get("PORT").Int(8080)
			srv := &http.Server{
				Addr:    ":" + strconv.Itoa(port),
				Handler: router,
			}
			options := make([]routes.Option, 0, 2)
			if envar.Get("HEALTH_CHECK").Bool(opts.HealthCheck) {
				options = append(options, routes.WithHealthCheck("/"))
			}
			if opts.BasicAuth != "" || envar.Get("BASIC_AUTH_ENABLE").Bool(opts.BasicAuthEnabled) {
				options = append(options, routes.WithBasicAuth("/api", opts.BasicAuth))
			}
			if err := routes.Setup(router, opts.DirPath, options...); err != nil {
				return err
			}
			defer routes.Shutdown()
			go func() {
				// service connections
				if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
					log.Fatal(ctx).Msgf("listen: %+v", fmt.Errorf("%w", err))
				}
			}()
			// Wait for interrupt signal to gracefully shutdown the server with
			// a timeout of 5 seconds.
			quit := make(chan os.Signal, 1)
			signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
			<-quit
			log.Info(ctx).Msgf("Shutdown Server ...(%v)", time.Now())
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := srv.Shutdown(ctx); err != nil {
				log.Fatal(ctx).Msgf("Failed Graceful Shutdown : %+v", fmt.Errorf("%w", err))
			}
			log.Info(ctx).Msgf("Server exiting. (%v)", time.Now())
			return nil
		},
	}
	cmd.Flags().StringVarP(&opts.DirPath, "dir", "d", "", "データディレクトリパス")
	cmd.Flags().BoolVarP(&opts.HealthCheck, "health", "H", false, "ヘルスチェックを有効にする")
	cmd.Flags().StringVarP(&opts.BasicAuth, "basic", "b", "", "Basic認証ユーザーパスワードを設定する")
	cmd.Flags().BoolVarP(&opts.BasicAuthEnabled, "basic-auth", "B", false, "Basic認証を有効にする")
	return cmd
}
