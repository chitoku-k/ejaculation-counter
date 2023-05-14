FROM rabbitmq:3.11.15-management
ARG RABBITMQ_MESSAGE_DEDUPLICATION_VERSION=0.5.3
ARG RABBITMQ_MESSAGE_DEDUPLICATION_ELIXIR_VERSION=1.12.2
ADD https://github.com/noxdafox/rabbitmq-message-deduplication/releases/download/$RABBITMQ_MESSAGE_DEDUPLICATION_VERSION/elixir-$RABBITMQ_MESSAGE_DEDUPLICATION_ELIXIR_VERSION.ez /opt/rabbitmq/plugins/
ADD https://github.com/noxdafox/rabbitmq-message-deduplication/releases/download/$RABBITMQ_MESSAGE_DEDUPLICATION_VERSION/rabbitmq_message_deduplication-$RABBITMQ_MESSAGE_DEDUPLICATION_VERSION.ez /opt/rabbitmq/plugins/
RUN chown rabbitmq:rabbitmq /opt/rabbitmq/plugins/*.ez && \
    rabbitmq-plugins enable --offline rabbitmq_message_deduplication
