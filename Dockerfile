FROM nginx:1.26.0
COPY conf /etc/nginx/templates
COPY public /var/www/html
