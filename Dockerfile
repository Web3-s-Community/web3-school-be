# syntax=docker/dockerfile:1

FROM golang:1.18-alpine

ENV GOPATH /app
ENV CGO_ENABLED 1

ARG SCHEDULER_ENABLED
ENV SCHEDULER_ENABLED=$SCHEDULER_ENABLED

WORKDIR /app/src/autopilot-helper
RUN mkdir bin data log pid

COPY go.mod go.sum ./
COPY helper/ ./helper
RUN apk add --update --no-cache gcc g++
RUN go mod tidy
RUN go build -o ./bin/autopilot-helper helper/main.go

COPY template/run.sh.template run.sh
# COPY helper/.env .
RUN chmod +x run.sh

EXPOSE 8092

ENTRYPOINT ["./run.sh", "start"]

