FROM nginx:1.21.1-alpine
COPY conf /etc/nginx/templates
COPY public /var/www/html
