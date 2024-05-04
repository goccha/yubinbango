package cmd

import (
	"context"
	"github.com/goccha/envar"
	"github.com/goccha/yubinbango/internal/routes"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
	ginlog "github.com/goccha/logging/gin"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(Lambda)
}

var Lambda = NewLambdaFunction()

// NewLambdaFunction LambdaFunction初期化
func NewLambdaFunction() *cobra.Command {
	type Options struct {
		Version string
	}
	options := &Options{}
	cmd := &cobra.Command{
		Use:   "lambda",
		Short: "lambda",
		Long:  "lambda",
		RunE: func(cmd *cobra.Command, args []string) error {
			switch options.Version {
			case "v2":
				lambda.Start(newV2Func())
			default:
				lambda.Start(newV1Func())
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&options.Version, "version", "v", "v1", "")
	return cmd
}

// newV2Func HttpAPI用LambdaFunction
func newV2Func() func(ctx context.Context, event events.APIGatewayV2HTTPRequest) (res events.APIGatewayV2HTTPResponse, err error) {
	var ginLambda *ginadapter.GinLambdaV2
	if router, err := ginNew(); err != nil {
		panic(err)
	} else {
		ginLambda = ginadapter.NewV2(router)
	}
	return func(ctx context.Context, event events.APIGatewayV2HTTPRequest) (res events.APIGatewayV2HTTPResponse, err error) {
		if err = logRequest(ctx, event, event.Body); err != nil {
			return
		}
		return ginLambda.ProxyWithContext(ctx, event)
	}
}

// newV1Func RestAPI用LambdaFunction
func newV1Func() func(ctx context.Context, event events.APIGatewayProxyRequest) (res events.APIGatewayProxyResponse, err error) {
	var ginLambda *ginadapter.GinLambda
	if router, err := ginNew(); err != nil {
		panic(err)
	} else {
		ginLambda = ginadapter.New(router)
	}
	return func(ctx context.Context, event events.APIGatewayProxyRequest) (res events.APIGatewayProxyResponse, err error) {
		if err = logRequest(ctx, event, event.Body); err != nil {
			return
		}
		return ginLambda.ProxyWithContext(ctx, event)
	}
}

// ginNew ginの初期化
func ginNew() (router *gin.Engine, err error) {
	router = gin.New()
	router.Use(ginlog.AccessLog(), gin.Recovery())
	err = routes.Setup(router, envar.String("DATA_DIR_PATH"), routes.WithHealthCheck("/"), routes.WithBasicAuth("/api", ""))
	return
}
