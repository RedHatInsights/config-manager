FROM registry.access.redhat.com/ubi8/go-toolset as builder
COPY --chown=default . .
ENV CGO_ENABLED=0
RUN ["go", "build", "-o", "config_manager", "."]

FROM registry.access.redhat.com/ubi8-minimal
MAINTAINER jassteph@redhat.com
COPY ./playbooks ./playbooks
COPY --from=builder /opt/app-root/src/config_manager .
ENTRYPOINT ["./config_manager"]
CMD ["run"]
