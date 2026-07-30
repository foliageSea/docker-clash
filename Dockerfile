# syntax=docker/dockerfile:1.7
FROM node:20-alpine AS web
WORKDIR /src/web
COPY web/package*.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

FROM golang:1.22-alpine AS app
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=web /src/cmd/webdist ./cmd/webdist
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/nexus ./cmd

FROM alpine:3.22 AS core
ARG TARGETARCH
ARG MIHOMO_VERSION=v1.19.28
RUN apk add --no-cache ca-certificates curl gzip \
    && case "$TARGETARCH" in \
      amd64) ASSET="mihomo-linux-amd64-compatible-${MIHOMO_VERSION}.gz"; SHA="70d01cfb8cb7bf7a92fd1af16cb4b9553d90bb4eecde3b5c4849103e27c80ddb" ;; \
      arm64) ASSET="mihomo-linux-arm64-${MIHOMO_VERSION}.gz"; SHA="2474450cd1c41dfa53036a54a4e85579f493d3af524d86c3d4b8e2b240b56cd2" ;; \
      *) echo "unsupported architecture: $TARGETARCH"; exit 1 ;; \
    esac \
    && curl -fL "https://github.com/MetaCubeX/mihomo/releases/download/${MIHOMO_VERSION}/${ASSET}" -o /tmp/mihomo.gz \
    && if [ "$MIHOMO_VERSION" = "v1.19.28" ]; then echo "$SHA  /tmp/mihomo.gz" | sha256sum -c -; fi \
    && gzip -dc /tmp/mihomo.gz > /mihomo \
    && chmod 0755 /mihomo

FROM alpine:3.22
RUN apk add --no-cache ca-certificates tzdata && addgroup -S nexus && adduser -S -G nexus -h /app nexus
WORKDIR /app
COPY --from=app /out/nexus /app/nexus
COPY --from=core /mihomo /app/bin/mihomo
RUN mkdir /data && chown nexus:nexus /data
USER nexus
ENV NEXUS_DATA_DIR=/data MIHOMO_BINARY=/app/bin/mihomo NEXUS_LISTEN=0.0.0.0:9080 GIN_MODE=release
EXPOSE 9080 7890
VOLUME ["/data"]
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 CMD wget -qO- http://127.0.0.1:9080/api/status >/dev/null || exit 1
ENTRYPOINT ["/app/nexus"]
