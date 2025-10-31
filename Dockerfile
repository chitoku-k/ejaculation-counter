FROM nginx:1.29.3
COPY conf /etc/nginx/templates
COPY public /var/www/html
