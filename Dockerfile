FROM nginx:1.25.2-alpine
COPY conf /etc/nginx/templates
COPY public /var/www/html
