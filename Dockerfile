FROM nginx:1.27.5
COPY conf /etc/nginx/templates
COPY public /var/www/html
