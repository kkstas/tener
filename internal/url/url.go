package url

import (
	"context"
	"slices"
	"strings"

	"github.com/awslabs/aws-lambda-go-api-proxy/core"
)

func Create(ctx context.Context, parts ...string) string {
	stageName := getStageName(ctx)
	return buildURL(stageName, parts...)
}

func buildURL(stageName string, parts ...string) string {
	if stageName != "" {
		parts = slices.Insert(parts, 0, stageName)
	}

	return "/" + strings.Join(parts, "/")
}

func getStageName(ctx context.Context) string {
	apiGWContext, ok := core.GetAPIGatewayContextFromContext(ctx)
	if ok {
		return apiGWContext.Stage
	}
	return ""
}
