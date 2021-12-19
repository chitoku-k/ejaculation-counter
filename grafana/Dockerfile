FROM grafana/grafana:7.5.12
COPY . /etc/grafana/provisioning/
RUN grafana-cli --pluginsDir="$GF_PATHS_PLUGINS" plugins install neocat-cal-heatmap-panel
