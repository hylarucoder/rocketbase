ARG GO_VERSION=1.21.5
ARG ALPINE_VERSION=3.17

# build stage
FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS build-env
RUN apk add build-base

ADD . /src

# CGO_ENABLED=0
ENV CGO_ENABLED=0
RUN cd /src && go build -tags pq -o server ./examples/base/main.go

# final stage
FROM alpine:${ALPINE_VERSION}

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
WORKDIR /app

COPY --from=build-env /src/server /app/

# create user and give it permissions
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
RUN mkdir /app/pb_data
RUN mkdir /app/pb_migrations
RUN mkdir /app/pb_config

RUN chown -R appuser:appgroup /app
USER appuser

CMD ["/app/server","serve", "--http=0.0.0.0:8090"]