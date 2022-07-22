FROM nginx:1.23.1-alpine
COPY conf /etc/nginx/templates
COPY public /var/www/html
