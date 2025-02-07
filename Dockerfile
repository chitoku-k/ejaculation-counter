FROM nginx:1.27.4
COPY conf /etc/nginx/templates
COPY public /var/www/html
