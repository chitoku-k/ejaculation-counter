FROM nginx:1.25.5-alpine
COPY conf /etc/nginx/templates
COPY public /var/www/html
