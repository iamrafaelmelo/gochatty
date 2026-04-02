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
npm install          # install frontend dependencies
npm run build        # build React + Tailwind static assets once
npm run dev          # run Vite dev server (optional)
go run ./cmd/main.go # run the Go server against built static files

# Or use the existing Make target for the Go app
make run
```

## Setup for Production environment

The production image builds frontend assets, compiles the Go binary in a separate stage, and ships only the binary plus static assets in a minimal `scratch` runtime image.

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
- Optional frontend override: set `APP_WEBSOCKET_URL` at build time (for example `wss://chat.example.com/ws`) to control websocket endpoint resolution.
- The current root `Dockerfile` is intended for production-style builds and deployment.
