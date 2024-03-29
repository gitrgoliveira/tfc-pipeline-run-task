# Use a minimal base image
FROM golang:1.21-alpine AS builder

# Set the working directory
WORKDIR /app

# Copy the Go module files
COPY go.mod ./

# Download the Go module dependencies
RUN go mod download && go mod verify

# Copy the source code
COPY . ./

# Build the Go binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o tfc-pipeline-run-task .

# Use a minimal base image for runtime
FROM alpine:latest

# Set the working directory
WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/tfc-pipeline-run-task ./

# Add terraform cli default v1.7.4
ARG VERSION=1.7.4
RUN apk add --update --virtual .deps --no-cache gnupg && \
    cd /tmp && \
    wget https://releases.hashicorp.com/terraform/${VERSION}/terraform_${VERSION}_linux_amd64.zip && \
    wget https://releases.hashicorp.com/terraform/${VERSION}/terraform_${VERSION}_SHA256SUMS && \
    wget https://releases.hashicorp.com/terraform/${VERSION}/terraform_${VERSION}_SHA256SUMS.sig && \
    wget -qO- https://www.hashicorp.com/.well-known/pgp-key.txt | gpg --import && \
    gpg --verify terraform_${VERSION}_SHA256SUMS.sig terraform_${VERSION}_SHA256SUMS && \
    grep terraform_${VERSION}_linux_amd64.zip terraform_${VERSION}_SHA256SUMS | sha256sum -c && \
    unzip /tmp/terraform_${VERSION}_linux_amd64.zip -d /tmp && \
    mv /tmp/terraform /usr/local/bin/terraform && \
    rm -f /tmp/terraform_${VERSION}_linux_amd64.zip terraform_${VERSION}_SHA256SUMS ${VERSION}/terraform_${VERSION}_SHA256SUMS.sig && \
    apk del .deps

COPY ./scripts ./scripts
# Set the executable permissions for the binary
RUN chmod +x ./tfc-pipeline-run-task

# Expose the port that the web service listens on
EXPOSE 80

# Set the entrypoint command
CMD ["./tfc-pipeline-run-task", "serve"]
