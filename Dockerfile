FROM golang:1.26.0-alpine3.23 as build

ARG version

RUN if [ -z "$version" ]; then \
			echo "version is not set"; \
			exit 1; \
    fi

RUN apk update && \
		apk add git

WORKDIR /go/src

COPY go.mod go.sum ./
RUN go mod download && \
		go mod verify

COPY . .

RUN go build  \
    -o ../bin/sync-watch-server \
    -ldflags "-X github.com/matthiasharzer/sync-watch-server/cmd/version.version=$version"  \
    ./main.go

FROM alpine:3.23

COPY --from=build /go/bin/sync-watch-server /usr/local/bin/sync-watch-server

WORKDIR /var/lib/hka-2fa-proxy

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/sync-watch-server"]

