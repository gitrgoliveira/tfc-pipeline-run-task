package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	rt "tfc-pipeline-run-task/runTask"
)

// rest of your code

type Job struct {
	Payload rt.RunTaskPayload
	Path    string
}

// Queue to store the jobs (JSON payloads)
var jobQueue = make(chan Job, 200)

func handlePayload(payload rt.RunTaskPayload, filename string, shasum string) error {
	log.Printf("Received filename: %s\n", filename)
	log.Printf("Received shasum: %s\n", shasum)
	log.Printf("Received payload: %v\n", payload)

	// Check if file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return fmt.Errorf("ERROR: File %s does not exist", filename)

	} // Ensure it's only targeting files under the "scripts" path
	if strings.HasPrefix(filename, "scripts") {
		return fmt.Errorf("ERROR: path must start with `scripts`")

	} // and prevent escape an attempt
	if strings.Contains(filename, "..") {
		return fmt.Errorf("ERROR: path cannot contain `..`")
	}

	// if it exists, check it matches the shasum256
	matches, err := FileMatchesSHA256(filename, strings.Split(shasum, ":")[1])
	if err != nil {
		return fmt.Errorf("ERROR when calculating hash %s", err.Error())
	}
	// if it doesn't, return 400 Bad Request
	if !matches {
		message := fmt.Sprintf("File %s does not match the checksum", filename)
		log.Println(message)
		// rt.SendError(message, payload)
		return fmt.Errorf(message)
	}

	log.Printf("File %s exists and matches the checksum\n", filename)

	if payload.Stage == "test" {
		log.Println("Test payload received. Job will not be aded to the queue")
		return nil
	}

	// Create a new job with the payload and the path
	job := Job{
		Payload: payload,
		Path:    filename,
	}

	// Add the job to the queue
	jobQueue <- job

	return nil
}

func LocalHandleRequest(w http.ResponseWriter, r *http.Request) {

	// Check if the request method is POST
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Parse the JSON payload
	var payload rt.RunTaskPayload
	err = json.Unmarshal(body, &payload)
	if err != nil {
		message := fmt.Sprintf("Error parsing JSON payload: %s", err.Error())
		http.Error(w, message, http.StatusBadRequest)
		log.Println(message)
		return
	}

	filename := fmt.Sprintf(".%s", r.URL.Path)
	values := r.URL.Query()
	shasum256 := values.Get("shasum")

	err = handlePayload(payload, filename, shasum256)
	if err != nil {
		log.Printf("%s", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Respond with an HTTP 200 OK status
	okResponse(w)
}

func okResponse(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("200 OK"))
}

func main() {
	// Start the job processor in a separate goroutine
	go processJobs()

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "serve":
			// Define the HTTP handler function
			http.HandleFunc("/", LocalHandleRequest)

			// Start the server on port 80
			log.Println("Server listening on port 80...")

			log.Fatal(http.ListenAndServe(":80", nil))
		}
	}
}

func processJobs() {
	for job := range jobQueue {
		payload := job.Payload

		log.Printf("Processing %s %+v\n", payload.Stage, payload.RunID)
		// Add your job processing logic here
		workspace := fmt.Sprintf("/tmp/%s-%s", payload.RunID, payload.Stage)
		// Create the workspace directory
		err := os.MkdirAll(workspace, os.ModePerm)
		if err != nil {
			log.Println(err.Error())
			continue
		}

		if payload.ConfigurationVersionDownloadURL != "" {
			err := rt.DownloadConfigVersion(payload.ConfigurationVersionDownloadURL, payload.AccessToken, workspace)
			if err != nil {
				log.Println(err.Error())
			}
		}

		// write the paylod to a file
		payloadFilePath := fmt.Sprintf("%s/payload.json", workspace)
		payloadFile, err := os.Create(payloadFilePath)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		// Convert the payload to JSON
		jsonData, err := json.Marshal(payload)
		if err != nil {
			log.Println(err.Error())
			return
		}

		// Write the payload to the file
		_, err = payloadFile.Write(jsonData)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		payloadFile.Close()

		// Download the plan file from the url in job.Payload.PlanJsonAPIURL
		if payload.PlanJsonAPIURL == "" {
			log.Println("No plan file to download")
		} else {
			planFilePath := fmt.Sprintf("%s/plan.json", workspace)
			err = rt.DownloadPlan(payload.PlanJsonAPIURL, payload.AccessToken, planFilePath)
			if err != nil {
				log.Println(err.Error())
				continue
			}
		}

		// Run the bash script
		results, err := RunCommands(workspace, job.Path, job.Payload)
		if err != nil {
			log.Println(err.Error())
			continue
		}

		errorFound := false

		// Check if any result is failed
		for _, outcome := range results {
			if outcome.Attributes.Tags.Status != nil {
				for _, tag := range outcome.Attributes.Tags.Status {
					if tag.Level == "error" {
						log.Println("Found at least one result with an error tag")
						errorFound = true
						break
					}
				}
				if errorFound {
					break
				}
			}

		}
		var finalPayload rt.RunTaskResult
		if errorFound {
			log.Println("At least one command failed")
			// Create a failed result with the results from all commands that were run.
			finalPayload = rt.CreateFailedResult("At least one command failed", results)
		} else {
			log.Println("All commands passed")
			finalPayload = rt.CreatePassedResult("All commands passed", results)
		}

		jsonResultData, err := json.Marshal(finalPayload)
		if err != nil {
			log.Fatalf("Failed to marshal final payload: %v", err)
			continue
		}
		// For debugging
		log.Println("Sending Payload: ", string(jsonResultData))
		err = rt.SendPatchRequest(payload.TaskResultCallbackURL, jsonResultData, payload.AccessToken)
		if err != nil {
			log.Println(err.Error())
			continue
		}

		os.RemoveAll(workspace)
	}
}
