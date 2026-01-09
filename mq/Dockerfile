FROM ubuntu:latest AS rabbitmq-message-deduplication
ARG RABBITMQ_MESSAGE_DEDUPLICATION_VERSION=0.7.3
ARG RABBITMQ_MESSAGE_DEDUPLICATION_FILENAME=plugins-rmqv4.2.x-erl27-elx1.18.zip
ADD https://github.com/noxdafox/rabbitmq-message-deduplication/releases/download/$RABBITMQ_MESSAGE_DEDUPLICATION_VERSION/$RABBITMQ_MESSAGE_DEDUPLICATION_FILENAME /usr/src/
RUN --mount=type=cache,id=mq:/var/cache/apt,target=/var/cache/apt \
    --mount=type=cache,id=mq:/var/lib/apt/lists,target=/var/lib/apt/lists \
    apt-get -y update && \
    apt-get -y install \
        unzip && \
    mkdir -p /opt/rabbitmq/plugins && \
    unzip -d /opt/rabbitmq/plugins "/usr/src/$RABBITMQ_MESSAGE_DEDUPLICATION_FILENAME"

FROM rabbitmq:4.2.2-management
COPY --from=rabbitmq-message-deduplication /opt/rabbitmq/plugins /opt/rabbitmq/plugins
RUN chown rabbitmq:rabbitmq /opt/rabbitmq/plugins/*.ez && \
    rabbitmq-plugins enable --offline rabbitmq_message_deduplication
