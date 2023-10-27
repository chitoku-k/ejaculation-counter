FROM nginx:1.25.3-alpine
COPY conf /etc/nginx/templates
COPY public /var/www/html
