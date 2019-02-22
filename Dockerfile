FROM golang:1.11 AS builder

RUN mkdir /app
WORKDIR /app

ADD ./go.mod /app/go.mod
ADD ./go.sum /app/go.sum
RUN go mod tidy

ADD ingress_rules_editor.go /app
RUN go build -o ingress_rules_editor ./ingress_rules_editor.go

# Runner
FROM gcr.io/cloud-builders/kubectl@sha256:abeaf7bd496e66301f92fd884683feaec6894bfb12d1a02fdbc98a920fab4968

LABEL "name"="ingress-rules-editor"
LABEL "version"="0.0.1"
LABEL "maintainer"="Masashi Shibata <shibata_masashi@cyberagent.co.jp>"

LABEL "com.github.actions.name"="GitHub Action to edit kubernetes ingress rules."
LABEL "com.github.actions.description"="Edit kubernetes ingress rules."
LABEL "com.github.actions.icon"="upload-cloud"
LABEL "com.github.actions.color"="green"

ENV DOCKERVERSION=18.06.1-ce
RUN apt-get update && apt-get -y --no-install-recommends install curl \
  && curl -fsSLO https://download.docker.com/linux/static/stable/x86_64/docker-${DOCKERVERSION}.tgz \
  && tar xzvf docker-${DOCKERVERSION}.tgz --strip 1 \
                 -C /usr/local/bin docker/docker \
  && rm docker-${DOCKERVERSION}.tgz \
  && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/ingress_rules_editor /builder/ingress_rules_editor

COPY entrypoint.sh /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
