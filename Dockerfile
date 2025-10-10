FROM nginx:1.29.2
COPY conf /etc/nginx/templates
COPY public /var/www/html
