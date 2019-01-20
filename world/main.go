package main

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// Response is of type APIGatewayProxyResponse since we're leveraging the
// AWS Lambda Proxy Request functionality (default behavior)
//
// https://serverless.com/framework/docs/providers/aws/events/apigateway/#lambda-proxy-integration
type Response events.APIGatewayProxyResponse

// Handler is our lambda handler invoked by the `lambda.Start` function call
func Handler(ctx context.Context, request events.APIGatewayProxyRequest) (Response, error) {

	// //Get the path parameter that was sent
	// name := request.PathParameters["name"]

	// //Generate message that want to be sent as body
	// message := fmt.Sprintf(" { \"Message\" : \"Hello %s \" } ", name)

	// //Returning response with AWS Lambda Proxy Response
	// return events.APIGatewayProxyResponse{Body: message, StatusCode: 200}, nil

	var buf bytes.Buffer

	body, err := json.Marshal(map[string]interface{}{
		"message": "Okay so your other function also executed successfully!",
	})
	if err != nil {
		return Response{StatusCode: 404}, err
	}
	json.HTMLEscape(&buf, body)

	resp := Response{
		StatusCode:      200,
		IsBase64Encoded: false,
		Body:            buf.String(),
		Headers: map[string]string{
			"Content-Type":           "application/json",
			"X-MyCompany-Func-Reply": "world-handler",
		},
	}

	return resp, nil
}

func main() {
	lambda.Start(Handler)
}
