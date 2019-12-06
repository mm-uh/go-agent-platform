# First stage
## Building
FROM golang:1.12.6-alpine3.10 as builder
WORKDIR /go/src/github.com/mm-uh/go-agent-platform
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o /main . 

# Second stage
## Minimization
FROM alpine:latest
COPY --from=builder /main /
RUN chmod +x /main
ENTRYPOINT ["/main"]
