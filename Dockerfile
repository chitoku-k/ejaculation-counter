FROM nginx:1.23.3-alpine
COPY conf /etc/nginx/templates
COPY public /var/www/html
