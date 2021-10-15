package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/terraform"
	test_structure "github.com/gruntwork-io/terratest/modules/test-structure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pinger/go-multicloud-deploy/src/function/v2"
)

// An example of how to test the Terraform module in examples/terraform-aws-lambda-example using Terratest.
func TestTerraformAwsLambdaExample(t *testing.T) {
	t.Parallel()

	// Make a copy of the terraform module to a temporary directory. This allows running multiple tests in parallel
	// against the same terraform module.
	exampleFolder := test_structure.CopyTerraformFolderToTemp(t, "../../", "infrastructure/aws/")

	// Give this lambda function a unique ID for a name so we can distinguish it from any other lambdas
	// in your AWS account
	functionName := fmt.Sprintf("terratest-aws-lambda-example-%s", random.UniqueId())

	// Pick a random AWS region to test in. This helps ensure your code works in all regions.
	awsRegion := aws.GetRandomStableRegion(t, nil, nil)

	// Construct the terraform options with default retryable errors to handle the most common retryable errors in
	// terraform testing.
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		// The path to where our Terraform code is located
		TerraformDir: exampleFolder,

		// Variables to pass to our Terraform code using -var options
		Vars: map[string]interface{}{
			"function_name": functionName,
			"region":        awsRegion,
		},
	})

	// At the end of the test, run `terraform destroy` to clean up any resources that were created
	defer terraform.Destroy(t, terraformOptions)

	// This will run `terraform init` and `terraform apply` and fail the test if there are any errors
	terraform.InitAndApply(t, terraformOptions)

	// Invoke the function, so we can test its output
	response := aws.InvokeFunction(t, awsRegion, functionName, function.Event{Code: 123, Message: "hi!"})

	// This function just echos it's input as a JSON string when `ShouldFail` is `false``
	assert.Equal(t, `"hi!"`, string(response))

	// Invoke the function, this time causing it to error and capturing the error
	_, err := aws.InvokeFunctionE(t, awsRegion, functionName, function.Event{Code: 0, Message: "hi!"})

	// Function-specific errors have their own special return
	functionError, ok := err.(*aws.FunctionError)
	require.True(t, ok)

	// Make sure the function-specific error comes back
	assert.Contains(t, string(functionError.Payload), "Failed to handle")
}

// Annother example of how to test the Terraform module in
// examples/terraform-aws-lambda-example using Terratest, this time with
// the aws.InvokeFunctionWithParams.
func TestTerraformAwsLambdaWithParamsExample(t *testing.T) {
	t.Parallel()

	// Make a copy of the terraform module to a temporary directory. This allows running multiple tests in parallel
	// against the same terraform module.
	exampleFolder := test_structure.CopyTerraformFolderToTemp(t, "../../", "infrastructure/aws/")

	// Give this lambda function a unique ID for a name so we can distinguish it from any other lambdas
	// in your AWS account
	functionName := fmt.Sprintf("terratest-aws-lambda-withparams-example-%s", random.UniqueId())

	// Pick a random AWS region to test in. This helps ensure your code works in all regions.
	awsRegion := aws.GetRandomStableRegion(t, nil, nil)

	// Construct the terraform options with default retryable errors to handle the most common retryable errors in
	// terraform testing.
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		// The path to where our Terraform code is located
		TerraformDir: exampleFolder,

		// Variables to pass to our Terraform code using -var options
		Vars: map[string]interface{}{
			"function_name": functionName,
			"region":        awsRegion,
		},
	})

	// At the end of the test, run `terraform destroy` to clean up any resources that were created
	defer terraform.Destroy(t, terraformOptions)

	// This will run `terraform init` and `terraform apply` and fail the test if there are any errors
	terraform.InitAndApply(t, terraformOptions)

	// Call InvokeFunctionWithParms with an InvocationType of "DryRun".
	// A "DryRun" invocation does not execute the function, so the example
	// test function will not be checking the payload.
	var invocationType aws.InvocationTypeOption = aws.InvocationTypeDryRun
	input := &aws.LambdaOptions{InvocationType: &invocationType}
	out := aws.InvokeFunctionWithParams(t, awsRegion, functionName, input)

	// With "DryRun", there's no message in the output, but there is
	// a status code which will have a value of 204 for a successful
	// invocation.
	assert.Equal(t, int(*out.StatusCode), 204)

	// Invoke the function, this time causing the Lambda to error and
	// capturing the error.
	invocationType = aws.InvocationTypeRequestResponse
	input = &aws.LambdaOptions{
		InvocationType: &invocationType,
		Payload:        function.Event{Code: 0, Message: "hi!"},
	}
	out, err := aws.InvokeFunctionWithParamsE(t, awsRegion, functionName, input)

	// The Lambda executed, but should have failed.
	assert.Error(t, err, "Unhandled")

	// Make sure the function-specific error comes back
	assert.Contains(t, string(out.Payload), "Failed to handle")

	// Call InvokeFunctionWithParamsE with a LambdaOptions struct that has
	// an unsupported InvocationType.  The function should fail.
	invocationType = "Event"
	input = &aws.LambdaOptions{
		InvocationType: &invocationType,
		Payload:        function.Event{Code: 123, Message: "hi!"},
	}
	out, err = aws.InvokeFunctionWithParamsE(t, awsRegion, functionName, input)
	require.NotNil(t, err)
	assert.Contains(t, err.Error(), "LambdaOptions.InvocationType, if specified, must either be \"RequestResponse\" or \"DryRun\"")
}

// An example of how to test the Terraform module in examples/terraform-gcp-lambda-example using Terratest.
func TestTerraformGoogleCloudFunctionsExample(t *testing.T) {
	t.Parallel()

	// Make a copy of the terraform module to a temporary directory. This allows running multiple tests in parallel
	// against the same terraform module.
	exampleFolder := test_structure.CopyTerraformFolderToTemp(t, "../../", "infrastructure/gcp/")

	// Give this lambda function a unique ID for a name so we can distinguish it from any other lambdas
	// in your gcp account
	functionName := fmt.Sprintf("terratest-gcp-functions-example-%s", random.UniqueId())

	// Construct the terraform options with default retryable errors to handle the most common retryable errors in
	// terraform testing.
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		// The path to where our Terraform code is located
		TerraformDir: exampleFolder,

		// Variables to pass to our Terraform code using -var options
		Vars: map[string]interface{}{
			"function_name": functionName,
		},
	})

	// At the end of the test, run `terraform destroy` to clean up any resources that were created
	defer terraform.Destroy(t, terraformOptions)

	// This will run `terraform init` and `terraform apply` and fail the test if there are any errors
	terraform.InitAndApply(t, terraformOptions)

	triggerUrl := terraform.Output(t, terraformOptions, "trigger_url")

	assert.Contains(t, triggerUrl, functionName)

	// Invoke the function, so we can test its output
	e1 := function.Event{Code: 123, Message: "hi!"}
	response, _ := InvokeFunction(triggerUrl, e1)

	// This function just echos it's input as a JSON string when `ShouldFail` is `false``
	// Unmarshal string into structs.
	b := []byte(response)
	var e function.Event
	json.Unmarshal(b, &e)
	assert.Equal(t, e1, e)

	// Invoke the function, this time causing it to error and capturing the error
	//_, err := gcp.InvokeFunctionE(t, gcpRegion, functionName, function.Event{Code: 0, Message: "hi!"})

	// Function-specific errors have their own special return
	//functionError, ok := err.(*gcp.FunctionError)
	//require.True(t, ok)

	// Make sure the function-specific error comes back
	//assert.Contains(t, string(functionError.Payload), "Failed to handle")

}

func InvokeFunction(triggerUrl string, event function.Event) (string, error) {

	j, _ := json.Marshal(event)
	str := string(j)
	jsonData := []byte(str)

	request, _ := http.NewRequest("POST", triggerUrl, bytes.NewBuffer(jsonData))
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{}
	response, error := client.Do(request)
	if error != nil {
		panic(error)
		return "", error
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	return string(body), err
}
