package test

import (
	"fmt"
	"testing"

	"github.com/gruntwork-io/terratest/modules/gcp"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/terraform"
	test_structure "github.com/gruntwork-io/terratest/modules/test-structure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pinger/go-multicloud-deploy/src/function/v2"
)

// An example of how to test the Terraform module in examples/terraform-gcp-lambda-example using Terratest.
func TestTerraformGoogleCloudFunctionsExample(t *testing.T) {
	t.Parallel()

	// Make a copy of the terraform module to a temporary directory. This allows running multiple tests in parallel
	// against the same terraform module.
	exampleFolder := test_structure.CopyTerraformFolderToTemp(t, "../../", "infrastructure/gcp/")

	// Give this lambda function a unique ID for a name so we can distinguish it from any other lambdas
	// in your gcp account
	functionName := fmt.Sprintf("terratest-gcp-functions-example-%s", random.UniqueId())

	// Pick a random gcp region to test in. This helps ensure your code works in all regions.
	gcpRegion := gcp.GetRandomRegion(t, nil, nil)

	// Construct the terraform options with default retryable errors to handle the most common retryable errors in
	// terraform testing.
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		// The path to where our Terraform code is located
		TerraformDir: exampleFolder,

		// Variables to pass to our Terraform code using -var options
		Vars: map[string]interface{}{
			"function_name": functionName,
			"region":        gcpRegion,
		},
	})

	// At the end of the test, run `terraform destroy` to clean up any resources that were created
	defer terraform.Destroy(t, terraformOptions)

	// This will run `terraform init` and `terraform apply` and fail the test if there are any errors
	terraform.InitAndApply(t, terraformOptions)

	// Invoke the function, so we can test its output
	response := gcp.InvokeFunction(t, gcpRegion, functionName, function.Event{Code: 123, Message: "hi!"})

	// This function just echos it's input as a JSON string when `ShouldFail` is `false``
	assert.Equal(t, `"hi!"`, string(response))

	// Invoke the function, this time causing it to error and capturing the error
	_, err := gcp.InvokeFunctionE(t, gcpRegion, functionName, function.Event{Code: 0, Message: "hi!"})

	// Function-specific errors have their own special return
	functionError, ok := err.(*gcp.FunctionError)
	require.True(t, ok)

	// Make sure the function-specific error comes back
	assert.Contains(t, string(functionError.Payload), "Failed to handle")
}

// Annother example of how to test the Terraform module in
// examples/terraform-gcp-lambda-example using Terratest, this time with
// the gcp.InvokeFunctionWithParams.
func TestTerraformGoogleCloudFunctionsWithParamsExample(t *testing.T) {
	t.Parallel()

	// Make a copy of the terraform module to a temporary directory. This allows running multiple tests in parallel
	// against the same terraform module.
	exampleFolder := test_structure.CopyTerraformFolderToTemp(t, "../../", "infrastructure/gcp/")

	// Give this lambda function a unique ID for a name so we can distinguish it from any other lambdas
	// in your gcp account
	functionName := fmt.Sprintf("terratest-gcp-functions-withparams-example-%s", random.UniqueId())

	// Pick a random gcp region to test in. This helps ensure your code works in all regions.
	gcpRegion := gcp.GetRandomRegion(t, nil, nil)

	// Construct the terraform options with default retryable errors to handle the most common retryable errors in
	// terraform testing.
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		// The path to where our Terraform code is located
		TerraformDir: exampleFolder,

		// Variables to pass to our Terraform code using -var options
		Vars: map[string]interface{}{
			"function_name": functionName,
			"region":        gcpRegion,
		},
	})

	// At the end of the test, run `terraform destroy` to clean up any resources that were created
	defer terraform.Destroy(t, terraformOptions)

	// This will run `terraform init` and `terraform apply` and fail the test if there are any errors
	terraform.InitAndApply(t, terraformOptions)

	// Call InvokeFunctionWithParms with an InvocationType of "DryRun".
	// A "DryRun" invocation does not execute the function, so the example
	// test function will not be checking the payload.
	var invocationType gcp.InvocationTypeOption = gcp.InvocationTypeDryRun
	input := &gcp.LambdaOptions{InvocationType: &invocationType}
	out := gcp.InvokeFunctionWithParams(t, gcpRegion, functionName, input)

	// With "DryRun", there's no message in the output, but there is
	// a status code which will have a value of 204 for a successful
	// invocation.
	assert.Equal(t, int(*out.StatusCode), 204)

	// Invoke the function, this time causing the Lambda to error and
	// capturing the error.
	invocationType = gcp.InvocationTypeRequestResponse
	input = &gcp.LambdaOptions{
		InvocationType: &invocationType,
		Payload:        function.Event{Code: 0, Message: "hi!"},
	}
	out, err := gcp.InvokeFunctionWithParamsE(t, gcpRegion, functionName, input)

	// The Lambda executed, but should have failed.
	assert.Error(t, err, "Unhandled")

	// Make sure the function-specific error comes back
	assert.Contains(t, string(out.Payload), "Failed to handle")

	// Call InvokeFunctionWithParamsE with a LambdaOptions struct that has
	// an unsupported InvocationType.  The function should fail.
	invocationType = "Event"
	input = &gcp.LambdaOptions{
		InvocationType: &invocationType,
		Payload:        function.Event{Code: 123, Message: "hi!"},
	}
	out, err = gcp.InvokeFunctionWithParamsE(t, gcpRegion, functionName, input)
	require.NotNil(t, err)
	assert.Contains(t, err.Error(), "LambdaOptions.InvocationType, if specified, must either be \"RequestResponse\" or \"DryRun\"")
}
