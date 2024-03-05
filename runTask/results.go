package runTask

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
)

func CreatePassedResult(message string, outcomes []RunTaskOutcome) RunTaskResult {
	result := RunTaskResult{}
	result.Data.Type = "task-results"
	result.Data.Attributes.Status = "passed"
	result.Data.Attributes.Message = message
	result.Data.Relationships = &RunTaskResultRelationships{
		Outcomes: struct {
			Data []RunTaskOutcome `json:"data"`
		}{
			Data: outcomes,
		},
	}
	return result

}
func CreateFailedResult(message string, outcomes []RunTaskOutcome) RunTaskResult {
	result := RunTaskResult{}
	result.Data.Type = "task-results"
	result.Data.Attributes.Status = "failed"
	result.Data.Attributes.Message = message
	result.Data.Relationships = &RunTaskResultRelationships{
		Outcomes: struct {
			Data []RunTaskOutcome `json:"data"`
		}{
			Data: outcomes,
		},
	}
	return result
}

// createErrorResult is a function that creates a RunTaskResult with a failed status and an error message.
// It takes a string parameter 'message' which represents the error message.
// The function returns a RunTaskResult struct.
func createErrorResult(message string) RunTaskResult {
	result := RunTaskResult{}
	result.Data.Type = "task-results"
	result.Data.Attributes.Status = "failed"
	result.Data.Attributes.Message = message

	outcome := RunTaskOutcome{}
	outcome.Type = "task-result-outcomes"
	outcome.Attributes.OutcomeID = "PIPELINE-ERROR"
	outcome.Attributes.Description = message
	outcome.Attributes.Body = message
	outcome.Attributes.Tags.Status = []RunTaskRichLabel{
		{
			Label: "Status",
			Level: "error",
		},
	}

	result.Data.Relationships = &RunTaskResultRelationships{
		Outcomes: struct {
			Data []RunTaskOutcome `json:"data"`
		}{
			Data: []RunTaskOutcome{outcome},
		},
	}

	return result
}

func SendPatchRequest(url string, payload []byte, authToken string) error {
	req, err := http.NewRequest(http.MethodPatch, url, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/vnd.api+json")
	req.Header.Set("Authorization", "Bearer "+authToken)

	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil || resp == nil {
		return err
	}

	defer resp.Body.Close()
	// Read the response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Convert the response body to a string
	respBodyStr := string(respBody)

	// Process the response as needed
	if resp.StatusCode != http.StatusOK {
		log.Println(respBodyStr)
		return err
	}

	return nil
}

func SendError(message string, payload RunTaskPayload) error {
	result := createErrorResult(message)
	jsonResultData, err := json.Marshal(result)
	if err != nil {
		return err
	}

	err = SendPatchRequest(payload.TaskResultCallbackURL, jsonResultData, payload.AccessToken)
	if err != nil {
		return err
	}
	return nil
}
