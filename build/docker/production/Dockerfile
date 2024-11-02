# Stage: Builder --------------------------------------------------------------
FROM golang:latest AS builder

WORKDIR /build

COPY go.mod go.sum .
RUN go mod download && go mod verify

COPY . .
RUN CGO_ENABLED=0 GOOS=linux \
    go build -C ./cmd -o ../bin/chat -installsuffix cgo

# Stage: Deploy ---------------------------------------------------------------
FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /opt/www/app

COPY --from=builder /build/bin .
EXPOSE 8080

CMD ["./chat"]
