# Frontend build stage
FROM node:20-alpine AS webbuild
WORKDIR /app/web
COPY web/package*.json ./
RUN npm install --no-audit --no-fund
COPY web/. .
RUN npm run build

# Backend build stage
FROM golang:1.22-alpine AS gobuild
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Copy built frontend into the expected static dir
COPY --from=webbuild /app/web/dist ./web/dist
# Build static binary
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o server ./main.go

# Final runtime image
FROM alpine:3.20
WORKDIR /app
RUN apk add --no-cache ca-certificates wget
COPY --from=gobuild /app/server /app/server
COPY --from=gobuild /app/web/dist /app/web/dist
ENV ADDR=:8080
EXPOSE 8080
HEALTHCHECK --interval=10s --timeout=3s --retries=5 CMD wget -q -O- http://localhost:8080/api/healthz || exit 1
CMD ["/app/server"]
