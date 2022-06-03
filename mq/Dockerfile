FROM rabbitmq:3.10.5-management
ADD https://github.com/noxdafox/rabbitmq-message-deduplication/releases/download/0.5.3/elixir-1.12.2.ez /opt/rabbitmq/plugins/
ADD https://github.com/noxdafox/rabbitmq-message-deduplication/releases/download/0.5.3/rabbitmq_message_deduplication-0.5.3.ez /opt/rabbitmq/plugins/
RUN chown rabbitmq:rabbitmq /opt/rabbitmq/plugins/*.ez && \
    rabbitmq-plugins enable --offline rabbitmq_message_deduplication
