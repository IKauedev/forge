# syntax=docker/dockerfile:1

# ---- build: instala dependências e gera o bundle de produção ----
FROM node:24-alpine AS builder
WORKDIR /app

COPY package.json package-lock.json ./
RUN --mount=type=cache,target=/root/.npm npm ci

COPY . .
RUN npm run build

# ---- runtime: Nginx sem privilégios, porta 8080 ----
FROM nginxinc/nginx-unprivileged:1.31-alpine AS runtime

COPY docker/nginx.conf /etc/nginx/conf.d/default.conf
COPY --chmod=755 docker/40-config.sh /docker-entrypoint.d/40-config.sh
COPY --from=builder --chown=101:101 /app/dist/starter-app/browser /usr/share/nginx/html

EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s \
  CMD wget -qO- http://127.0.0.1:8080/ >/dev/null || exit 1
