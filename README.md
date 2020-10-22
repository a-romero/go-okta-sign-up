# go-sign-up

AWS lambda function that processes sign-up requests via Okta and updates the internal tables
accordingly to complete the onboarding process.

## Setup
The Okta events are captured via Event Hooks, as explained:
https://developer.okta.com/docs/concepts/event-hooks/

Once configured against the url, the Event Hook requires an initial GET request against the
service, passing a `challenge` via the `X-Okta-Verification-Challenge` header, and expects
it back as a json object:

```
{
  "verification": "challengeString"
}
```

The above serves as a way of validating the service configured is legitimate. Once verified,
the GET method is not used again.


Okta Group Rules are defined as part of the setup too. These rules allow to automatically
add a newly signed up user to a Group based on their email domain. This way, once a user
signs up by introducing their details into the Okta sign-up form, it will generate an event
for that user that will then get picked up by this service.

The event will look like:

```
{
    "eventType": "com.okta.event_hook",
    ...
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
                    "displayName": "Test Test"
                },
                "outcome": {
                    "result": "SUCCESS"
                },
                "target": [
                    {
                        "id": "00u10r4h1xAe2T675as4x7",
                        "type": "User",
                        "alternateId": "test@test.com",
                        "displayName": "A Test"
                    },
                    {
                        "id": "00g10rdyamywSK8Pw4x7",
                        "type": "UserGroup",
                        "alternateId": "unknown",
                        "displayName": "EXT-TestClient"
                    }
                ],
                ...
            }
        ]
    },
    "eventTime": "2020-09-21T12:58:52.580Z",
    "contentType": "application/json"
}
```

## Service Event Process
The service will unmarshal the event, and:

1. Check if the `Client` for the new user already exists
2. If it does not exist it will create a record for the `Client`
3. Create `User` record extracting the details from the Okta event

Response from the service is always expected to be `200` with an empty body.

## AWS Resources

### DynamoDB Tables

- Client
- User

## Build
```
make build
```

## Deploy to AWS Lambda
```
sls deploy
```
