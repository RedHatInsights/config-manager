FROM registry.access.redhat.com/ubi8/go-toolset as builder
WORKDIR /go/src/app
COPY . .
ENV CGO_ENABLED=0
USER root
RUN ["go", "build", "-o", "config_manager", "."]
USER 1001

FROM registry.access.redhat.com/ubi8-minimal
MAINTAINER jassteph@redhat.com
COPY ./playbooks ./playbooks
COPY --from=builder /go/src/app/config_manager .
ENTRYPOINT ["./config_manager"]
CMD ["run"]
