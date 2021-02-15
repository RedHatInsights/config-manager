FROM registry.redhat.io/ubi8/go-toolset
MAINTAINER jassteph@redhat.com

COPY ./config_manager ./config_manager
COPY ./db/migrations ./db/migrations

ENTRYPOINT ["./config_manager"]