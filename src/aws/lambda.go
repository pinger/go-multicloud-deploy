package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/pinger/go-multicloud-deploy/src/function/v2"
)

// Fails if ShouldFail is `true`, otherwise echos the input.
func HandleRequest(ctx context.Context, evnt function.Event) (string, error) {
	if evnt.Code == 0 {
		return "", fmt.Errorf("Failed to handle %#v", evnt)
	}
	return evnt.Message, nil
}

func main() {
	lambda.Start(HandleRequest)
}
