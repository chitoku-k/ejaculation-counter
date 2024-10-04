FROM nginx:1.27.2
COPY conf /etc/nginx/templates
COPY public /var/www/html
