# TFC Pipeline RunTask

This is a run task for Terraform Cloud/Enterprise that allows for the implementation of customizable scripts/commands at various stages in Terraform's workflow, thus enforcing them across all projects or workspaces. This feature provides an efficient and streamlined method to introduce scripts into your Terraform deployment pipeline.

The different run task stages supported are:
 * pre-plan
 * post-plan
 * pre-apply
 * post-apply

**Due to the missing TLS in the listener, please do not use this project as is.**
 
## Getting Started
These instructions will help you run the project on your local machine for development and testing purposes.
The project is designed to be run in a containerized environment, so you can easily deploy it to any cloud provider or local machine.

### Prerequisites

 * Docker
 * Python3

### Running
These are the steps you need to follow to get this project up and running on your local machine for development and testing purposes.

1. Clone this repository to your local machine.
2. Navigate into the project directory.
3. Run the docker compose file with: `docker compose up --build`
4. The application should now be accessible at http://localhost.

### Usage
This project exposes an HTTP server on port 80. It accepts POST requests with a JSON payload that matches the RunTaskPayload struct, and then processes the payload from a job queue.

The path in the url needs to match the path to an existing script along with its sha256 sum.
To test the application is running, you can `POST` a test payload.

```bash
$ curl -X POST "http://localhost/scripts/pre_plan.sh?shasum=sha256:e72d4cbd289cae46769aa8302f9bb3f34858f60370175104bcfbe213bbadb468" -d @test_payload.json
```

You can customize the scripts by changing the scripts in the `scripts` directory. The script names should match the ones specified in the RunTaskPayload struct.

To retrieve the `shasum` parameter you can run this command and add the output to the URL as `?shasum=sha256:`
```bash
$ shasum -a 256 ./scripts/pre_apply.sh
e72d4cbd289cae46769aa8302f9bb3f34858f60370175104bcfbe213bbadb468  ./scripts/pre_apply.sh
```

The scripts are then run in a unique folder that contains:
 * A `payload.json` file with the contentss of the received JSON payload.
 * The uploaded configuration version - i.e. all terraform files received by TFC/E
 * If there's a plan file, it's downloaded to a file called `plan.json`

### Setting up your Run Task for testing
If you are just experimenting in your own laptop you need to:
 * Run `docker compose up --build`
 * Test with `curl -X POST "http://0.0.0.0/scripts/pre_plan.sh?shasum=sha256:e72d4cbd289cae46769aa8302f9bb3f34858f60370175104bcfbe213bbadb468" -d @test_payload.json` or similar
 * Discover your external IP address
 * Ensure  your firewall allows incoming traffic on port 80
 * Run the same `curl` test with your external IP.
 * Create a new Run Task with the following settings:
    * Endpoint URL: `http://<your_external_ip>:80/scripts/pre_plan.sh?shasum=sha256:e72d4cbd289cae46769aa8302f9bb3f34858f6`

## Note on Security
This project does not include TLS encryption as it's a minimum requirement for production. In a real-world scenario, you should consider adding TLS encryption (like Let's Encrypt) to ensure secure communication between clients and the server.

For security, the code is limited to only run scripts in the `scripts` directory. 
