FROM nginx:1.21.6-alpine
COPY conf /etc/nginx/templates
COPY public /var/www/html
