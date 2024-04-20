FROM nginx:1.25.5
COPY conf /etc/nginx/templates
COPY public /var/www/html
