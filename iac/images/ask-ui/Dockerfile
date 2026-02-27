FROM nginx:1.28.0-alpine-slim
LABEL authors="vikazeni"

COPY /ask-ui/dist /usr/share/nginx/html
COPY /iac/images/ask-ui/nginx.conf /etc/nginx/conf.d/default.conf