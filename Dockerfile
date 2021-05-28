FROM nginx:1.21.0-alpine
COPY conf /etc/nginx/templates
COPY public /var/www/html
