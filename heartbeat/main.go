package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// Response is of type APIGatewayProxyResponse since we're leveraging the
// AWS Lambda Proxy Request functionality (default behavior)
//
// https://serverless.com/framework/docs/providers/aws/events/apigateway/#lambda-proxy-integration
type Response events.APIGatewayProxyResponse

const expiration = time.Hour

var client = &http.Client{}

// Handler is our lambda handler invoked by the `lambda.Start` function call
func Handler(ctx context.Context, request events.APIGatewayProxyRequest) (Response, error) {

	apiKey := request.QueryStringParameters["api_key"]

	// Ensure all fields are not empty
	if apiKey != "" {

		log.Printf("Info: Request API key %s", apiKey)

		message := fmt.Sprintf(" { \"status\" : \"%s\" } ", "success")

		//Returning response with AWS Lambda Proxy Response
		return Response{StatusCode: 200,
			Body: message,
			Headers: map[string]string{
				"Access-Control-Allow-Origin":      "*",
				"Access-Control-Allow-Credentials": "true",
			},
		}, nil

	}

	// Missing one of required parameters
	log.Printf("Error: Request missing a required parameter")
	return Response{StatusCode: 400,
		Headers: map[string]string{
			"Access-Control-Allow-Origin":      "*",
			"Access-Control-Allow-Credentials": "true",
		},
	}, nil
}

type LocalServer struct{}

func (l *LocalServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		w.WriteHeader(500)
		w.Write([]byte(fmt.Sprintf("Failed to write body: %v", err)))
		return
	}

	url, err := url.Parse(r.URL.String())
	if err != nil {
		log.Printf("Error parsing query string: %v", err)
		w.WriteHeader(500)
		w.Write([]byte(fmt.Sprintf("Malformed query string: %v", err)))
		return
	}
	queryParams := url.Query()

	//**building request**
	req := events.APIGatewayProxyRequest{
		Body:                  string(body),
		Headers:               make(map[string]string),
		HTTPMethod:            r.Method,
		Path:                  r.URL.Path,
		QueryStringParameters: make(map[string]string),
	}

	//map raw request headers
	for k, v := range r.Header {
		req.Headers[strings.ToLower(k)] = v[0]
	}

	//Map raw query params
	for k, v := range queryParams {
		req.QueryStringParameters[strings.ToLower(k)] = v[0]
	}

	resp, err := Handler(r.Context(), req)
	if err != nil {
		log.Printf("Error handling request: %v", err)
		w.WriteHeader(500)
		w.Write([]byte(fmt.Sprintf("Error handling request: %v", err)))
		return
	}
	for k, v := range resp.Headers {
		w.Header().Add(k, v)
	}
	(w).Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(resp.StatusCode)
	w.Write([]byte(resp.Body))
}

func local() {
	server := &LocalServer{}
	fmt.Println("Starting local dev server on :8080")
	http.ListenAndServe(":8080", server)
}

func main() {
	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") == "" {
		//see local creds file for env vars
		local()
	} else {
		// Make the handler available for Remote Procedure Call by AWS Lambda
		lambda.Start(Handler)
	}
}
