FROM golang:1.11 AS builder

ENV GO111MODULE=on
RUN mkdir /app
WORKDIR /app

ADD ./go.mod /app/go.mod
ADD ./go.sum /app/go.sum
ADD ./main.go /app/main.go
RUN go build -o ingress_rules_editor ./main.go

# Runner
FROM alpine:latest

RUN apk --update add ca-certificates atop

LABEL "name"="ingress-rules-editor"
LABEL "version"="0.0.2"
LABEL "maintainer"="Masashi Shibata <shibata_masashi@cyberagent.co.jp>"

LABEL "com.github.actions.name"="GitHub Action to edit kubernetes ingress rules."
LABEL "com.github.actions.description"="Edit kubernetes ingress rules."
LABEL "com.github.actions.icon"="upload-cloud"
LABEL "com.github.actions.color"="green"

COPY entrypoint.sh /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]

