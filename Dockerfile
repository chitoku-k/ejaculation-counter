FROM nginx:1.27.3
COPY conf /etc/nginx/templates
COPY public /var/www/html
