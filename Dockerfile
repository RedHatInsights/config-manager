FROM registry.redhat.io/ubi8/go-toolset as builder
WORKDIR /go/src/app
COPY . .
ENV CGO_ENABLED=0
USER root
RUN ["go", "build", "-o", "config_manager", "main.go"]
USER 1001

FROM registry.redhat.io/ubi8-minimal
MAINTAINER jassteph@redhat.com
COPY ./db/migrations ./db/migrations
COPY --from=builder /go/src/app/config_manager .
ENTRYPOINT ["./config_manager"]
