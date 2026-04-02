FROM node:22-alpine AS assets

WORKDIR /src

COPY package.json package-lock.json ./
COPY vite.config.js ./
COPY tailwind.config.js postcss.config.js ./
COPY web ./web

RUN npm ci && npm run build

FROM golang:1.25.8-alpine3.23 AS builder

WORKDIR /src

ARG TARGETOS=linux
ARG TARGETARCH=amd64

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags="-s -w" -o /out/chat ./cmd/main.go

FROM scratch

WORKDIR /app

COPY --from=builder /out/chat /app/chat
COPY --from=assets /src/web/public /app/web/public

ENV APP_PORT=8080

EXPOSE 8080

USER 65532:65532

ENTRYPOINT ["/app/chat"]
