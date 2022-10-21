FROM nginx:1.23.2-alpine
COPY conf /etc/nginx/templates
COPY public /var/www/html
