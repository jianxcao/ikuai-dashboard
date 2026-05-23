FROM node:24-alpine AS frontend-build
WORKDIR /src/frontend
ARG VITE_APP_VERSION=dev
ENV VITE_APP_VERSION=${VITE_APP_VERSION}
RUN corepack enable
COPY frontend/package.json frontend/pnpm-lock.yaml frontend/pnpm-workspace.yaml ./
RUN pnpm install --frozen-lockfile --ignore-scripts
COPY frontend/ ./
RUN pnpm run build

FROM golang:1.26-alpine AS backend-build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY backend/ ./backend/
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/ikuai-dashboard ./backend/cmd/server

FROM alpine:3.22
WORKDIR /app
RUN apk add --no-cache ca-certificates dumb-init gosu shadow tzdata \
	&& addgroup -S -g 10001 ikuai \
	&& adduser -S -D -H -u 10001 -G ikuai -s /sbin/nologin ikuai \
	&& mkdir -p /config /defaults \
	&& chown -R ikuai:ikuai /app /config
COPY --from=backend-build /out/ikuai-dashboard /app/ikuai-dashboard
COPY --from=frontend-build /src/frontend/dist /app/frontend/dist
COPY config.example.yaml /defaults/config.example.yaml
COPY --chmod=755 docker/entrypoint.sh /usr/local/bin/docker-entrypoint.sh
ENV IKUAI_CONFIG=/config/config.yaml \
	PUID=1000 \
	PGID=1000 \
	UMASK=022
EXPOSE 8080
VOLUME ["/config"]
ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]
CMD ["/app/ikuai-dashboard"]
