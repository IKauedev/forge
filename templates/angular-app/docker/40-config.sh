#!/bin/sh
# Gera o config.json de runtime a partir de variáveis de ambiente (a imagem é a mesma em todo ambiente).
# Sem API_BASE_URL, mantém o config.json que veio no build.
set -eu

if [ -n "${API_BASE_URL:-}" ]; then
  cat > /usr/share/nginx/html/config.json <<EOF
{
  "apiBaseUrl": "${API_BASE_URL}",
  "auth": {
    "enabled": ${AUTH_ENABLED:-false},
    "loginUrl": "${AUTH_LOGIN_URL:-/login}"
  }
}
EOF
fi
