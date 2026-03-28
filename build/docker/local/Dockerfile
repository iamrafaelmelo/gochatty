FROM golang:1.25.8-alpine3.23

ENV PATH=/usr/local/go/bin:/usr/local/bin:${PATH}
ENV GOLANGCI_LINT_VERSION=v2.11.4
ENV GOVULNCHECK_VERSION=v1.1.4

RUN set -ex \
    && apk upgrade && apk update \
    && apk add build-base git make wget \
    && apk add --update nodejs npm \
    && wget -O- -nv https://golangci-lint.run/install.sh | sh -s -- -b /usr/local/bin ${GOLANGCI_LINT_VERSION} \
    && GOBIN=/usr/local/bin go install golang.org/x/vuln/cmd/govulncheck@${GOVULNCHECK_VERSION}

WORKDIR /opt/www/app

ENTRYPOINT ["tail", "-f", "/dev/null"]
