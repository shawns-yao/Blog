FROM alpine:3.21 AS og-fonts

RUN apk add --no-cache fontconfig

COPY deploy/fonts/ /usr/share/fonts/og/
RUN fc-cache -f

FROM node:22-alpine AS builder

WORKDIR /app

RUN corepack enable

COPY web/package.json web/pnpm-lock.yaml web/pnpm-workspace.yaml ./
RUN pnpm install --frozen-lockfile

COPY web/. .
COPY shared /shared

ARG APP_VERSION=dev
ARG BUILD_COMMIT=unknown
ENV APP_VERSION=${APP_VERSION} \
    BUILD_COMMIT=${BUILD_COMMIT}

RUN pnpm build

FROM node:22-alpine AS runtime

WORKDIR /app

ARG APP_VERSION=dev
ARG BUILD_COMMIT=unknown
ENV NODE_ENV=production \
    APP_VERSION=${APP_VERSION} \
    BUILD_COMMIT=${BUILD_COMMIT}

RUN apk add --no-cache su-exec \
    && addgroup -g 10001 -S app \
    && adduser -u 10001 -S app -G app \
    && corepack enable

COPY web/package.json web/pnpm-lock.yaml web/pnpm-workspace.yaml ./
RUN pnpm install --frozen-lockfile --prod

COPY --from=builder /app/build /app/build

# Install the bundled OTF/TTF fonts for OG rendering without release downloads.
COPY --from=og-fonts /usr/share/fonts/og/ /usr/share/fonts/og/
ARG ALPINE_MIRROR=https://dl-cdn.alpinelinux.org/alpine
RUN sed -i "s#https\?://dl-cdn.alpinelinux.org/alpine#${ALPINE_MIRROR%/}#g" /etc/apk/repositories \
    && apk add --no-cache fontconfig && fc-cache -f

COPY deploy/docker/renderer-entrypoint.sh /usr/local/bin/renderer-entrypoint.sh
RUN chmod +x /usr/local/bin/renderer-entrypoint.sh

EXPOSE 3000

ENTRYPOINT ["/usr/local/bin/renderer-entrypoint.sh"]
