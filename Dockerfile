FROM nginx:1.29.0
COPY conf /etc/nginx/templates
COPY public /var/www/html
