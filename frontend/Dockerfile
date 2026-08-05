# Dev stage
FROM node:24-alpine AS dev
WORKDIR /app
COPY package*.json ./
RUN npm ci
CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0"]

# Build stage
FROM node:24-alpine AS builder
WORKDIR /build
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

# Production stage
FROM nginx:1.29-alpine
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /build/dist /usr/share/nginx/html

RUN chown -R 101:101 /var/cache/nginx && \
    chown -R 101:101 /var/log/nginx && \
    touch /var/run/nginx.pid && chown 101:101 /var/run/nginx.pid

# numeric uid, not the "nginx" name: kubelet cannot verify a non-numeric user
# against runAsNonRoot and refuses to start the container
USER 101

EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://127.0.0.1:8080/ || exit 1

CMD ["nginx", "-g", "daemon off;"]
