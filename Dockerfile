FROM nginx:1.17.9-alpine
COPY conf /etc/nginx/conf.d
COPY public /var/www/html
CMD ["/bin/ash", "-c", "sed -i \"s/reactor:/$REACTOR_HOST:/;s/grafana:/$GF_HOST:/\" /etc/nginx/conf.d/default.conf && nginx -g 'daemon off;'"]