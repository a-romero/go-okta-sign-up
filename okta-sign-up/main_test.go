package main

import (
	"testing"
  "net/http"
  "encoding/json"

	"github.com/stretchr/testify/assert"
  "github.com/stretchr/testify/mock"
  "github.com/aws/aws-lambda-go/events"
)

type mockSignerUp struct {
  mock.Mock
}

func (su *SignerUp) TestGETMissingHeaderRequest(t *testing.T) {

	testEvent := events.APIGatewayProxyRequest{
    HTTPMethod: "GET",
    Headers: map[string]string{
      "BadTest": "BadTest",
    },
  }

  resp, _ := su.signUp(testEvent)

  assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

}

func (su *SignerUp) TestGETCorrectRequest(t *testing.T) {

	testEvent := events.APIGatewayProxyRequest{
    HTTPMethod: "GET",
    Headers: map[string]string{
      "X-Okta-Verification-Challenge": "TestChallenge",
    },
  }

  resp, err := su.signUp(testEvent)

  testChallenge := testEvent.Headers["X-Okta-Verification-Challenge"]
  testVerification := Verification {
    Verification: testChallenge,
  }

  testJs, err := json.Marshal(testVerification)
  if err != nil {
    errorLogger.Printf("%s", err)
  }

  assert.Nil(t, err)
  assert.Equal(t, testJs, []byte(resp.Body))
}

func (su *SignerUp) TestPOSTUnknownEventRequest(t *testing.T) {

	testEvent := events.APIGatewayProxyRequest{
    HTTPMethod: "POST",
    Body: "BadTest: badtest",
  }

  resp, _ := su.signUp(testEvent)

  assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

}

func (su *SignerUp) TestPOSTCorrectEventRequest(t *testing.T) {

	testEvent := events.APIGatewayProxyRequest{
    HTTPMethod: "POST",
    Body:
      `{
            "eventType": "com.okta.event_hook",
            "eventTypeVersion": "1.0",
            "cloudEventsVersion": "0.1",
            "source": "https://.okta.com/api/v1/eventHooks/who10jf3dotVCGgaj4x7",
            "eventId": "33b5bedc-ea7a-4503-aade-4b1aac0c465e",
            "data": {
                "events": [
                    {
                        "uuid": "2d050e48-fc0a-11ea-bbf6-1d98632c66b8",
                        "published": "2020-09-21T12:58:42.340Z",
                        "eventType": "group.user_membership.add",
                        "version": "0",
                        "displayMessage": "Add user to group membership",
                        "severity": "INFO",
                        "client": {
                            "ipChain": []
                        },
                        "actor": {
                            "id": "00u2yuxqc4rUOtf3J4x6",
                            "type": "User",
                            "alternateId": "test@test.com",
                            "displayName": "Test User"
                        },
                        "outcome": {
                            "result": "SUCCESS"
                        },
                        "target": [
                            {
                                "id": "00u10r4h1xAe2TIcx4x7",
                                "type": "User",
                                "alternateId": "test@test.com",
                                "displayName": "Test User"
                            },
                            {
                                "id": "00g10rdyamywSK8Pw4x7",
                                "type": "UserGroup",
                                "alternateId": "unknown",
                                "displayName": "EXT-TestGroup"
                            }
                        ],
                        "transaction": {
                            "type": "JOB",
                            "id": "eru10rdj52DWefteH4x7",
                            "detail": {}
                        },
                        "debugContext": {
                            "debugData": {
                                "threatSuspected": "false",
                                "targetEventHookIds": "who10jf3dotVCGgaj4x7"
                            }
                        },
                        "legacyEventType": "core.user_group_member.user_add",
                        "authenticationContext": {
                            "authenticationStep": 0,
                            "externalSessionId": "trsZmgqz3lkS7ezRi_mij-E-Q"
                        },
                        "securityContext": {}
                    }
                ]
            },
            "eventTime": "2020-09-21T12:58:52.580Z",
            "contentType": "application/json"
        }
      `,
      }

  resp, _ := su.signUp(testEvent)

  assert.Equal(t, http.StatusOK, resp.StatusCode)

}
