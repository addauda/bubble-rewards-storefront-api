package main

import (
	"context"
	"database/sql"
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
	_ "github.com/lib/pq"
)

// Response is of type APIGatewayProxyResponse since we're leveraging the
// AWS Lambda Proxy Request functionality (default behavior)
//
// https://serverless.com/framework/docs/providers/aws/events/apigateway/#lambda-proxy-integration
type Response events.APIGatewayProxyResponse

const expiration = time.Hour

var client = &http.Client{}

func GenerateCouponCodeQuery() string {
	return "SELECT redemptions_coupon.id, submissions.instagram_account, rewards.description, redemptions_coupon.status from public.redemptions_coupon join submissions on redemptions_coupon.submission_id = submissions.id join offers on submissions.offer_id = offers.id join rewards on offers.loyalty_reward_id = rewards.id WHERE code = $1 AND redemptions_coupon.status = 'PENDING' AND current_timestamp < redemptions_coupon.expire_at"
}

func GenerateInstantQuery() string {
	return "SELECT submissions.id, submissions.instagram_account, rewards.description from submissions join offers on submissions.offer_id = offers.id join rewards on offers.instant_reward_id = rewards.id WHERE submissions.instagram_account = $1 AND submissions.status = 'ACCEPTED' AND current_timestamp < submissions.instant_reward_expire_at LIMIT 1"
}

// Handler is our lambda handler invoked by the `lambda.Start` function call
func Handler(ctx context.Context, request events.APIGatewayProxyRequest) (Response, error) {

	code := request.QueryStringParameters["code"]
	redemptionType := request.QueryStringParameters["redemption_type"]
	apiKey := request.QueryStringParameters["api_key"]

	// Ensure all fields are not empty
	if code != "" && redemptionType != "" && apiKey != "" {

		log.Printf("Info: Request code %s", code)
		log.Printf("Info: Request redemption type %s", redemptionType)
		log.Printf("Info: Request API key %s", apiKey)

		// Connect to database
		connStr := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
			os.Getenv("DB_HOST"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"))

		db, err := sql.Open("postgres", connStr)
		if err != nil {
			log.Printf("Error: %v", err)
			return Response{StatusCode: 500,
				Body: "{}",
				Headers: map[string]string{
					"Access-Control-Allow-Origin":      "*",
					"Access-Control-Allow-Credentials": "true",
				},
			}, nil
		}

		defer db.Close()

		// Validate API key
		var storeID int
		var storeName string
		row := db.QueryRow("SELECT id,name FROM stores WHERE api_key = $1", apiKey)
		switch err = row.Scan(&storeID, &storeName); err {
		case sql.ErrNoRows:
			log.Printf("Error: No store with API key [%s] was found", apiKey)
			return Response{StatusCode: 401,
				Headers: map[string]string{
					"Access-Control-Allow-Origin":      "*",
					"Access-Control-Allow-Credentials": "true",
				},
			}, nil
		case nil:
			log.Printf("Info: Retreived store as [%s]", storeName)
		default:
			log.Printf("Error: %v", err)
			return Response{StatusCode: 500,
				Headers: map[string]string{
					"Access-Control-Allow-Origin":      "*",
					"Access-Control-Allow-Credentials": "true",
				},
			}, nil
		}

		// Redeem a coupon code
		if redemptionType == "COUPON" {
			log.Printf("Info: Validating redemption type [%s]", redemptionType)
			var redemptionID int
			var instagramAccount string
			var rewardDescription string
			var redemptionStatus string
			row := db.QueryRow(GenerateCouponCodeQuery(), code)
			switch err = row.Scan(&redemptionID, &instagramAccount, &rewardDescription, &redemptionStatus); err {
			case sql.ErrNoRows:
				log.Printf("Error: Redemption code [%s] NOT FOUND", code)
				return Response{StatusCode: 404,
					Headers: map[string]string{
						"Access-Control-Allow-Origin":      "*",
						"Access-Control-Allow-Credentials": "true",
					},
				}, nil
			case nil:
				log.Printf("Success: Redemption code [%s] FOUND", code)

				//Generate message that want to be sent as body
				message := fmt.Sprintf(" { \"redemptionID\" : \"%d\", \"instagramAccount\" : \"%s\", \"rewardDescription\" : \"%s\", \"redemptionStatus\" : \"%s\", \"storeName\" : \"%s\" } ", redemptionID, instagramAccount, rewardDescription, redemptionStatus, storeName)

				//Returning response with AWS Lambda Proxy Response
				return Response{StatusCode: 200,
					Body: message,
					Headers: map[string]string{
						"Access-Control-Allow-Origin":      "*",
						"Access-Control-Allow-Credentials": "true",
					},
				}, nil

			default:
				log.Printf("Error: %v", err)
				return Response{StatusCode: 500,
					Headers: map[string]string{
						"Access-Control-Allow-Origin":      "*",
						"Access-Control-Allow-Credentials": "true",
					},
				}, nil
			}
		} else if redemptionType == "INSTANT" {
			log.Printf("Info: Validating redemption type [%s]", redemptionType)
			var submissionID int
			var instagramAccount string
			var rewardDescription string
			row := db.QueryRow(GenerateInstantQuery(), code)
			switch err = row.Scan(&submissionID, &instagramAccount, &rewardDescription); err {
			case sql.ErrNoRows:
				log.Printf("Error: Redemption code [%s] NOT FOUND", code)
				return Response{StatusCode: 404,
					Headers: map[string]string{
						"Access-Control-Allow-Origin":      "*",
						"Access-Control-Allow-Credentials": "true",
					},
				}, nil
			case nil:
				log.Printf("Success: Redemption code [%s] FOUND", code)

				//Generate message that want to be sent as body
				message := fmt.Sprintf(" { \"submissionId\" : \"%d\", \"instagramAccount\" : \"%s\", \"rewardDescription\" : \"%s\",  \"storeName\" : \"%s\" } ", submissionID, instagramAccount, rewardDescription, storeName)

				//Returning response with AWS Lambda Proxy Response
				return Response{StatusCode: 200,
					Body: message,
					Headers: map[string]string{
						"Access-Control-Allow-Origin":      "*",
						"Access-Control-Allow-Credentials": "true",
					},
				}, nil

			default:
				log.Printf("Error: %v", err)
				return Response{StatusCode: 500,
					Headers: map[string]string{
						"Access-Control-Allow-Origin":      "*",
						"Access-Control-Allow-Credentials": "true",
					},
				}, nil
			}
		} else {
			log.Printf("Error: Invalid redemption type [%s]", redemptionType)
			return Response{StatusCode: 400,
				Headers: map[string]string{
					"Access-Control-Allow-Origin":      "*",
					"Access-Control-Allow-Credentials": "true",
				},
			}, nil
		}
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
