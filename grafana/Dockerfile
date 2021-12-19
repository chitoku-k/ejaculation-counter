FROM grafana/grafana:8.3.3
COPY . /etc/grafana/provisioning/
RUN grafana-cli --pluginsDir="$GF_PATHS_PLUGINS" plugins install neocat-cal-heatmap-panel
