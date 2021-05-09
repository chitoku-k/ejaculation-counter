FROM nginx:1.19.10-alpine
COPY conf /etc/nginx/templates
COPY public /var/www/html
