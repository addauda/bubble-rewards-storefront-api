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

	//Get the path parameter that was sent
	activationCode := request.PathParameters["activation_code"]

	if activationCode != "" {
		cookie := http.Cookie{
			Name:    "activation_code",
			Value:   activationCode,
			Path:    "/",
			Expires: time.Now().Add(expiration),
		}

		resp := Response{
			StatusCode:      200,
			IsBase64Encoded: false,
			Body:            "{'hey':'there'}", //buf.String(),
			Headers:         map[string]string{"Set-Cookie": cookie.String()},
		}
		return resp, nil
	}

	return Response{StatusCode: 500}, nil

	// //Generate message that want to be sent as body
	// message := fmt.Sprintf(" { \"Message\" : \"Hello %s \" } ", name)

	// //Returning response with AWS Lambda Proxy Response
	// return events.APIGatewayProxyResponse{Body: message, StatusCode: 200}, nil

	// return events.APIGatewayProxyResponse{
	// 	StatusCode: 200,
	// 	Body:       "{}",
	// 	Headers:    map[string]string{"Set-Cookie": cookie.String()},
	// }, nil

	// var buf bytes.Buffer

	// body, err := json.Marshal(map[string]interface{}{
	// 	"message": "Okay so your other function also executed successfully!",
	// })
	// if err != nil {
	// 	return Response{StatusCode: 404}, err
	// }
	// json.HTMLEscape(&buf, body)

	// resp := Response{
	// 	StatusCode:      200,
	// 	IsBase64Encoded: false,
	// 	Body:            buf.String(),
	// 	Headers:         map[string]string{"Set-Cookie": cookie.String()},
	// }

	//return resp, nil
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
		Body:           string(body),
		Headers:        make(map[string]string),
		HTTPMethod:     r.Method,
		Path:           r.URL.Path,
		PathParameters: make(map[string]string),
	}

	//map raw request headers
	for k, v := range r.Header {
		req.Headers[strings.ToLower(k)] = v[0]
	}

	//Map raw query params
	for k, v := range queryParams {
		req.PathParameters[strings.ToLower(k)] = v[0]
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
		local()
	} else {
		// Make the handler available for Remote Procedure Call by AWS Lambda
		lambda.Start(Handler)
	}
}
