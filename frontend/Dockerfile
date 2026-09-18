# Build stage
FROM node:20-alpine AS builder

WORKDIR /build

COPY package*.json ./
RUN npm ci

COPY . .
RUN npm run build

# Runtime stage
FROM nginxinc/nginx-unprivileged:alpine

COPY --from=builder /build/dist /usr/share/nginx/html

RUN rm -f /etc/nginx/conf.d/default.conf && \
    echo 'server { \
        listen 8080; \
        server_name _; \
        root /usr/share/nginx/html; \
        index index.html; \
        location / { \
            try_files $uri $uri/ /index.html; \
        } \
        location /api/ { \
            proxy_pass http://access-tracker-backend:8080/; \
            proxy_set_header Host $host; \
            proxy_set_header X-Real-IP $remote_addr; \
        } \
        location /graphql { \
            proxy_pass http://access-tracker-backend:8080/graphql; \
            proxy_set_header Host $host; \
            proxy_set_header X-Real-IP $remote_addr; \
        } \
        location /oauth/ { \
            proxy_pass http://access-tracker-backend:8080/oauth/; \
            proxy_set_header Host $host; \
            proxy_set_header X-Real-IP $remote_addr; \
        } \
        location /healthz { \
            return 200 "healthy"; \
            add_header Content-Type text/plain; \
        } \
    }' > /etc/nginx/conf.d/default.conf

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/healthz || exit 1
