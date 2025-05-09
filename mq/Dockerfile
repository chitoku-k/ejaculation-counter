FROM rabbitmq:4.1.0-management
ARG RABBITMQ_MESSAGE_DEDUPLICATION_VERSION=
ARG RABBITMQ_MESSAGE_DEDUPLICATION_ELIXIR_VERSION=
ADD https://github.com/noxdafox/rabbitmq-message-deduplication/releases/download/$RABBITMQ_MESSAGE_DEDUPLICATION_VERSION/elixir-$RABBITMQ_MESSAGE_DEDUPLICATION_ELIXIR_VERSION.ez /opt/rabbitmq/plugins/
ADD https://github.com/noxdafox/rabbitmq-message-deduplication/releases/download/$RABBITMQ_MESSAGE_DEDUPLICATION_VERSION/rabbitmq_message_deduplication-$RABBITMQ_MESSAGE_DEDUPLICATION_VERSION.ez /opt/rabbitmq/plugins/
RUN chown rabbitmq:rabbitmq /opt/rabbitmq/plugins/*.ez && \
    rabbitmq-plugins enable --offline rabbitmq_message_deduplication
