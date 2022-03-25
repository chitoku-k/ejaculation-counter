FROM rabbitmq:3.9.14-management
ADD https://github.com/noxdafox/rabbitmq-message-deduplication/releases/download/0.5.2/elixir-1.12.2.ez /opt/rabbitmq/plugins/
ADD https://github.com/noxdafox/rabbitmq-message-deduplication/releases/download/0.5.2/rabbitmq_message_deduplication-0.5.2.ez /opt/rabbitmq/plugins/
RUN chown rabbitmq:rabbitmq /opt/rabbitmq/plugins/*.ez && \
    rabbitmq-plugins enable --offline rabbitmq_message_deduplication
