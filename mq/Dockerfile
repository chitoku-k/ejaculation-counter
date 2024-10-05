FROM rabbitmq:4.0.2-management
ARG RABBITMQ_MESSAGE_DEDUPLICATION_VERSION=0.6.3
ARG RABBITMQ_MESSAGE_DEDUPLICATION_ELIXIR_VERSION=1.16.3
ADD https://github.com/noxdafox/rabbitmq-message-deduplication/releases/download/$RABBITMQ_MESSAGE_DEDUPLICATION_VERSION/elixir-$RABBITMQ_MESSAGE_DEDUPLICATION_ELIXIR_VERSION.ez /opt/rabbitmq/plugins/
ADD https://github.com/noxdafox/rabbitmq-message-deduplication/releases/download/$RABBITMQ_MESSAGE_DEDUPLICATION_VERSION/rabbitmq_message_deduplication-$RABBITMQ_MESSAGE_DEDUPLICATION_VERSION.ez /opt/rabbitmq/plugins/
RUN chown rabbitmq:rabbitmq /opt/rabbitmq/plugins/*.ez && \
    rabbitmq-plugins enable --offline rabbitmq_message_deduplication
