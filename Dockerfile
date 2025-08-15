FROM nginx:1.29.1
COPY conf /etc/nginx/templates
COPY public /var/www/html
