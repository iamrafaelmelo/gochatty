# Go Chatty

![Preview](https://raw.githubusercontent.com/iamrafaelmelo/gochatty/refs/heads/master/screenshots/chat.png)

## Requirements

- Golang >= 1.18
- Docker
- Node & NPM
- Make

## Get started

```sh
make up
make container
```

## Building & running

All this following command below must be execute inside docker container.

```sh
npm run watch
make build
make run
```

And access `http://127.0.0.1:8080` on your browser.

## Environment

Copy `.env.example` and configure:

- `APP_NAME`
- `APP_PORT`
- `APP_URL`
- `APP_ALLOWED_ORIGINS` required, comma-separated origins allowed to open websocket connections
- `APP_ENV`
