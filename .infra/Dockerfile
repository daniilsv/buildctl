FROM golang:1.25-alpine AS backend-builder

WORKDIR /app

COPY back/go.mod back/go.sum* ./
RUN go mod download

COPY back/. .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server


FROM node:20-alpine AS frontend-builder

RUN corepack enable && corepack prepare pnpm@latest --activate

WORKDIR /app
COPY front/package.json front/pnpm-lock.yaml* ./
RUN pnpm install --frozen-lockfile

COPY front/. .
RUN pnpm run build


FROM nginx:alpine

RUN apk --no-cache add supervisor ca-certificates

WORKDIR /app

COPY --from=backend-builder /app/server .
COPY --from=backend-builder /app/db/migrations ./db/migrations

COPY --from=frontend-builder /app/dist /usr/share/nginx/html

COPY .infra/nginx.conf /etc/nginx/conf.d/default.conf
COPY .infra/supervisord.conf /etc/supervisord.conf

EXPOSE 80

CMD ["/usr/bin/supervisord", "-c", "/etc/supervisord.conf"]

