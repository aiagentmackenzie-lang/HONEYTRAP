FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
COPY cmd ./cmd
COPY internal ./internal
COPY profiles ./profiles
RUN GOCACHE=/tmp/gocache go build -o /out/honeytrap ./cmd/honeytrap

FROM alpine:3.21
RUN addgroup -S honeytrap && adduser -S honeytrap -G honeytrap
WORKDIR /srv/honeytrap
COPY --from=build /out/honeytrap /usr/local/bin/honeytrap
RUN mkdir -p /srv/honeytrap/var && chown -R honeytrap:honeytrap /srv/honeytrap
USER honeytrap
EXPOSE 2222 8080 2121 9161/udp
ENTRYPOINT ["honeytrap"]
CMD ["status"]
