package main

import (
	"fmt"
	"strings"

	"github.com/a-romero/go-dyndb/records"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

// Client ...
type Client struct {
	ClientID string `json:"clientId"`
}

// User ...
type User struct {
	ClientID string `json:"clientId"`
	UserEmail string `json:"userEmail"`
	ClientName string `json:"clientName"`
	IsAdmin bool `json:"isAdmin"`
	UserName string `json:"userName"`
}

// SignerUp ...
type SignerUp struct {
	clientTable string
  userTable string
}

// NewSignerUp ...
func NewSignerUp(clientTable, userTable string) *SignerUp {
	return &SignerUp{
    clientTable: clientTable,
		userTable: userTable,
	}
}

func (signerUp *SignerUp) writeToDB(event OktaEvent) error {

	var userArray []User

	var clientID, clientName, userEmail, userName string
	var isAdmin bool

	for _, target := range event.Data.Events[0].Targets {

		if target.Type == "User" {
			userEmail = target.AlternateID
			isAdmin = true
			userName = target.DisplayName
		}
		if target.Type == "UserGroup" {
			clientID = strings.Split(target.DisplayName, "-")[1]
			clientName = strings.Split(target.DisplayName, "-")[1]
		}
	}

	user := User{
		ClientID: clientID,
		ClientName: clientName,
		UserEmail: userEmail,
		IsAdmin: isAdmin,
		UserName: userName,
	}

	if !signerUp.clientExists(clientID) {
		var clientArray []Client
		client := Client{
			ClientID: clientID,
		}
		clientArray = append(clientArray, client)
		errWriteRecords := records.WriteRecord(clientArray, signerUp.clientTable, records.CreateDynDBSvc())
		if errWriteRecords != nil {
			fmt.Printf("Error writing records: %v", errWriteRecords)
			return errWriteRecords
		}
	}

	userArray = append(userArray, user)
	errWriteRecords := records.WriteRecord(userArray, signerUp.userTable, records.CreateDynDBSvc())
	if errWriteRecords != nil {
		fmt.Printf("Error writing records: %v", errWriteRecords)
		return errWriteRecords
	}

	return nil
}

func (signerUp *SignerUp) clientExists(clientID string) bool {

	var exists bool
	svc := records.CreateDynDBSvc()

	params := &dynamodb.QueryInput{
		TableName: aws.String(signerUp.clientTable),
		AttributesToGet: []*string{
			aws.String("clientId"),
		},
		KeyConditions: map[string]*dynamodb.Condition{
			"asset": &dynamodb.Condition{
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					&dynamodb.AttributeValue{
						S: aws.String(clientID),
					},
				},
			},
		},
		Limit:            aws.Int64(1),
		ScanIndexForward: aws.Bool(false),
	}

	result, err := svc.Query(params)
	if err != nil {
		return false
	}

	if len(result.Items) > 0 { exists = true}
	if len(result.Items) <= 0 { exists = false}

	return exists
}
