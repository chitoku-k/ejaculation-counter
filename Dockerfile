FROM nginx:1.23.4-alpine
COPY conf /etc/nginx/templates
COPY public /var/www/html
