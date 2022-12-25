FROM golang:alpine AS builder
LABEL maintainer="Mikolaj Gasior"

RUN apk add --update git bash openssh make gcc musl-dev

WORKDIR /go/src/github.com/mikogs/grafana-sidecar-backup-tool
COPY . .
RUN go build

FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /bin
COPY --from=builder /go/src/github.com/mikogs/grafana-sidecar-backup-tool/grafana-sidecar-backup-tool .

ENTRYPOINT ["/bin/grafana-sidecar-backup-tool"]
