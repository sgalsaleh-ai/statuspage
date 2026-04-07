# Stage 1: Build frontend
FROM node:22-alpine AS frontend
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN --mount=type=cache,target=/root/.npm npm ci
COPY frontend/ ./
RUN npm run build

# Stage 2: Build backend
FROM golang:1.25-alpine AS backend
WORKDIR /app
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY cmd/ ./cmd/
COPY internal/ ./internal/
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -o /statuspage ./cmd/server

# Stage 3: Final image
FROM alpine:3.20
ARG TARGETARCH
RUN apk add --no-cache ca-certificates curl
# Install support-bundle CLI for in-app bundle generation
RUN curl -sL https://github.com/replicatedhq/troubleshoot/releases/latest/download/support-bundle_linux_${TARGETARCH}.tar.gz \
    | tar xz -C /usr/local/bin support-bundle
WORKDIR /app
COPY --from=backend /statuspage ./statuspage
COPY --from=frontend /app/frontend/dist ./frontend/dist
ENV FRONTEND_DIR=/app/frontend/dist
EXPOSE 8080
ENTRYPOINT ["/app/statuspage"]
