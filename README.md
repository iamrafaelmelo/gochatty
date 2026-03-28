# Go Chatty

![Preview](https://raw.githubusercontent.com/iamrafaelmelo/gochatty/refs/heads/master/screenshots/chat.png)

> [!WARNING]
> This project is not production-ready, but contributions still are welcome.

## Requirements

- Go >=1.25
- Docker
- Node & NPM
- Make

## Local development

Install dependencies:

```sh
npm install          # install dependencies
npm run build        # build the CSS once
npm run watch        # run the Tailwind watcher
go run ./cmd/main.go # run the Go server

# Or just run all in background
npm run watch & go run ./cmd/main.go
# Or use the existing Make target
make run
```

## Setup for Production environment

The production image builds Tailwind CSS during the image build, compiles the Go binary in a separate stage, and ships only the binary plus static assets in a minimal `scratch` runtime image.

Build the image:

```sh
docker build -t gochatty .
```

Run the container:

```sh
docker run --rm -p 8080:8080 \
  -e APP_PORT=8080 \
  -e APP_ALLOWED_ORIGINS=http://0.0.0.0:8080 \
  -e APP_ENV=production \
  gochatty
```

Then access `http://0.0.0.0:8080`.

## Notes

- The production image runs as a non-root user.
- `APP_ALLOWED_ORIGINS` must include the browser origin that will open the websocket connection.
- The current root `Dockerfile` is intended for production-style builds and deployment.
