FROM grafana/grafana:7.5.16
COPY . /etc/grafana/provisioning/
RUN grafana-cli --pluginsDir="$GF_PATHS_PLUGINS" plugins install neocat-cal-heatmap-panel
