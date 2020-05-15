package dtclient

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEventDataMarshal(t *testing.T) {
	testJSONInput := []byte(`{
		"eventType": "MARKED_FOR_TERMINATION",
		"start": 20,
		"end": 20,
		"description": "K8s node was marked unschedulable. Node is likely being drained",
		"attachRules": {
			"entityIds": [ "HOST-CA78D78BBC6687D3" ]
		},
		"source": "OneAgent Operator"
	}`)

	var testEventData EventData
	err := json.Unmarshal(testJSONInput, &testEventData)
	assert.NoError(t, err)
	assert.Equal(t, testEventData.EventType, "MARKED_FOR_TERMINATION")
	assert.ElementsMatch(t, testEventData.AttachRules.EntityIDs, []string{"HOST-CA78D78BBC6687D3"})
	assert.Equal(t, testEventData.Source, "OneAgent Operator")

	jsonBuffer, err := json.Marshal(testEventData)
	assert.NoError(t, err)
	assert.JSONEq(t, string(jsonBuffer), string(testJSONInput))
}

func testSendEvent(t *testing.T, dynatraceClient Client) {
	{
		testValidEventData := []byte(`{
			"eventType": "MARKED_FOR_TERMINATION",
			"start": 20,
			"end": 20,
			"description": "K8s node was marked unschedulable. Node is likely being drained",
			"attachRules": {
				"entityIds": [ "HOST-CA78D78BBC6687D3" ]
			},
			"source": "OneAgent Operator"
		}`)
		var testEventData EventData
		err := json.Unmarshal(testValidEventData, &testEventData)
		assert.NoError(t, err)

		_, err = dynatraceClient.SendEvent(&testEventData)
		assert.NoError(t, err)
	}
	{
		testInvalidEventData := []byte(`{
			"start": 20,
			"end": 20,
			"description": "K8s node was marked unschedulable. Node is likely being drained",
			"attachRules": {
				"entityIds": [ "HOST-CA78D78BBC6687D3" ]
			},
			"source": "OneAgent Operator"
		}`)
		var testEventData EventData
		err := json.Unmarshal(testInvalidEventData, &testEventData)
		assert.NoError(t, err)

		_, err = dynatraceClient.SendEvent(&testEventData)
		assert.Error(t, err, "no eventType set")
	}
	{
		testExtraKeysEventData := []byte(`{
			"eventType": "MARKED_FOR_TERMINATION",
			"start": 20,
			"end": 20,
			"description": "K8s node was marked unschedulable. Node is likely being drained",
			"attachRules": {
				"entityIds": [ "HOST-CA78D78BBC6687D3" ]
			},
			"source": "OneAgent Operator",
		 	"cat": "potato"
		}`)
		var testEventData EventData
		err := json.Unmarshal(testExtraKeysEventData, &testEventData)
		assert.NoError(t, err)

		eid, err := dynatraceClient.SendEvent(&testEventData)
		assert.NoError(t, err)
		assert.Equal(t, &EventResponse{
			StoredEventIds:       []int64{42},
			StoredIds:            []string{"42"},
			StoredCorrelationIDs: []string{},
		}, eid)
	}
}

func handleSendEvent(request *http.Request, writer http.ResponseWriter) {
	eventPostResponse := []byte(`{
		"storedEventIds": [42],
		"storedIds": ["42"],
		"storedCorrelationIds": []
	}`)

	switch request.Method {
	case "POST":
		writer.WriteHeader(http.StatusOK)
		_, _ = writer.Write([]byte(eventPostResponse))
	default:
		writeError(writer, http.StatusMethodNotAllowed)
	}
}
