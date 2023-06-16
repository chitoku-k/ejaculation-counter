FROM nginx:1.25.1-alpine
COPY conf /etc/nginx/templates
COPY public /var/www/html
