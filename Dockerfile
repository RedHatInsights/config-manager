FROM golang:latest
WORKDIR /go/src/app
COPY . .
ENV CGO_ENABLED=0
RUN ["go", "build", "-o", "config_manager", "main.go"]

FROM registry.redhat.io/ubi8/go-toolset
MAINTAINER jassteph@redhat.com
COPY ./db/migrations ./db/migrations
COPY --from=0 /go/src/app/config_manager .
ENTRYPOINT ["./config_manager"]
