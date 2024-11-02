FROM golang:1.22-alpine3.19

RUN set -ex \
    && apk upgrade && apk update \
    && apk add make \
    && apk add --update nodejs npm

WORKDIR /opt/www/app

ENTRYPOINT ["tail", "-f", "/dev/null"]
