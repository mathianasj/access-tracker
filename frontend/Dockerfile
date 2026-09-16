# Build stage
FROM node:20-alpine AS builder

WORKDIR /build

COPY package*.json ./
RUN npm ci

COPY . .
RUN npm run build

# Runtime stage
FROM alpine:3.19

RUN apk add --no-cache nginx

RUN mkdir -p /run/nginx /var/cache/nginx /var/log/nginx /var/lib/nginx/tmp /etc/nginx/conf.d

COPY --from=builder /build/dist /usr/share/nginx/html

RUN echo 'daemon off;' > /etc/nginx/nginx.conf

RUN cat > /etc/nginx/httpd.conf <<'EOF'
worker_processes auto;
error_log /var/log/nginx/error.log warn;
pid /run/nginx.pid;

events {
    worker_connections 1024;
}

http {
    include /etc/nginx/mime.types;
    default_type application/octet-stream;

    log_format main '$remote_addr - $remote_user [$time_local] "$request" '
                      '$status $body_bytes_sent "$http_referer" '
                      '$http_user_agent';

    access_log /var/log/nginx/access.log main;
    sendfile on;
    tcp_nopush on;
    keepalive_timeout 65;

    server {
        listen 80;
        server_name _;
        root /usr/share/nginx/html;
        index index.html;

        location / {
            try_files $uri $uri/ /index.html;
        }

        location /api/ {
            proxy_pass http://access-tracker-backend:8080/;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
        }

        location /healthz {
            return 200 "healthy";
            add_header Content-Type text/plain;
        }
    }
}
EOF

RUN cat > /etc/nginx/conf.d/default.conf <<'EOF'
server {
    listen 8080;
    server_name _;
    root /usr/share/nginx/html;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location /api/ {
        proxy_pass http://access-tracker-backend:8080/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    location /healthz {
        return 200 "healthy";
        add_header Content-Type text/plain;
    }
}
EOF

EXPOSE 80 8080

RUN echo 'healthy' > /usr/share/nginx/html/healthz

CMD ["nginx", "-c", "/etc/nginx/httpd.conf"]
