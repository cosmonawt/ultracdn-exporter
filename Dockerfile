FROM golang
WORKDIR /go/src/github.com/cosmonawt/ultracdn-exporter
ADD . /go/src/github.com/cosmonawt/ultracdn-exporter
ENV GO111MODULE=on
ENV CGO_ENABLED=0
RUN go build -o exporter *.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=0 /go/src/github.com/cosmonawt/ultracdn-exporter/exporter /bin/ultracdn-exporter
ENTRYPOINT ["ultracdn-exporter"]