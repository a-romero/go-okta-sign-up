package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var stringRegexp = regexp.MustCompile(`[A-Za-z0-9]+`)
var intRegexp = regexp.MustCompile(`\d+`)

var errorLogger = log.New(os.Stderr, "ERROR ", log.Llongfile)

// Verification ...
type Verification struct {
	Verification string `json:"verification"`
}

// Target ...
type Target struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	AlternateID string `json:"alternateId"`
	DisplayName string `json:"displayName"`
}

// Outcome ...
type Outcome struct {
	Result string `json:"result"`
}

// Event ...
type Event struct {
	UUID           string                 `json:"uuid"`
	Published      string                 `json:"published"`
	EventType      string                 `json:"eventType"`
	Version        string                 `json:"version"`
	DisplayMessage string                 `json:"display"`
	Severity       string                 `json:"severity"`
	Actor          map[string]interface{} `json:"actor"`
	Outcome        Outcome                `json:"outcome"`
	Targets        []Target               `json:"target,omitempty"`
}

// Data ...
type Data struct {
	Events []Event `json:"events,omitempty"`
}

// OktaEvent ...
type OktaEvent struct {
	EventID   string `json:"eventId"`
	Data      Data   `json:"data,omitempty"`
	EventTime string `json:"eventTime"`
}

type Configuration struct {
	ClientTable         string `json:"clientTable"`
	UserTable string `json:"userTable"`
}

const (
	defaultConfigFilePath = "./config/config.json"
	// APP_ADD_USER to assign an app to a user
	APP_ADD_USER = "application.user_membership.add"
	// USER_CREATE to create a new user
	USER_CREATE = "user.lifecycle.create"
	// GROUP_ADD_USER to add a user to a group
	GROUP_ADD_USER = "group.user_membership.add"
)

func (signerUp *SignerUp) signUp(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	var js []byte
	var err error

	if req.HTTPMethod == "GET" {

		challenge := req.Headers["X-Okta-Verification-Challenge"]
		if challenge == "" {
			err := fmt.Errorf("no X-Okta-Verification-Challenge provided in request")
			errorLogger.Printf("%s", err)
			return clientError(http.StatusBadRequest)
		}

		verification := Verification{
			Verification: challenge,
		}

		js, err = json.Marshal(verification)
		if err != nil {
			errorLogger.Printf("%s", err)
			return serverError(err)
		}
	}

	if req.HTTPMethod == "POST" {
		var event OktaEvent
		if err := json.Unmarshal([]byte(req.Body), &event); err != nil {
			err := fmt.Errorf("unknown Okta event")
			errorLogger.Printf("%s", err)
			return clientError(http.StatusBadRequest)
		}

		if event.Data.Events[0].EventType == GROUP_ADD_USER {
			if err := signerUp.writeToDB(event); err != nil {
				errorLogger.Printf("%s", err)
				return serverError(err)
			}
		}
		fmt.Printf("%v", event)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Content-Type":                     "application/json",
			"Access-Control-Allow-Origin":      "*",
			"Access-Control-Allow-Credentials": "true",
			"Access-Control-Allow-Methods":     "OPTIONS,POST,GET",
		},
		Body: string(js),
	}, nil
}

func serverError(err error) (events.APIGatewayProxyResponse, error) {
	errorLogger.Println(err.Error())

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusInternalServerError,
		Body:       http.StatusText(http.StatusInternalServerError),
	}, nil
}

func clientError(status int) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		StatusCode: status,
		Body:       http.StatusText(status),
	}, nil
}

func getConfig() (*Configuration, error) {
	data, err := ioutil.ReadFile(defaultConfigFilePath)
	if err != nil {
		return nil, err
	}

	var cfg Configuration
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func main() {

	cfg, err := getConfig()
	if err != nil {
		log.Fatalf("Error getting config: %s", err)
	}

	log.Printf("Running lambda with config: %#v", cfg)

	signerUp := NewSignerUp(cfg.ClientTable, cfg.UserTable)

	lambda.Start(signerUp.signUp)
}
