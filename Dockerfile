FROM nginx:1.27.0
COPY conf /etc/nginx/templates
COPY public /var/www/html
