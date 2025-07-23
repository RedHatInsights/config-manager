FROM registry.access.redhat.com/ubi9/go-toolset:1.24.4-1753221510 as builder
COPY --chown=default . .
ENV CGO_ENABLED=0
RUN ["go", "build", "-o", "config_manager", "."]

FROM registry.access.redhat.com/ubi9-minimal:latest
COPY --from=builder /opt/app-root/src/config_manager .
ENTRYPOINT ["./config_manager"]
CMD ["run"]
