# Frontend build stage
FROM node:24-alpine3.22 AS webbuild
WORKDIR /app/web
COPY web/package*.json ./
RUN npm install --no-audit --no-fund
COPY web/. .
RUN npm run build

# Backend build stage
FROM golang:1.25-alpine3.22 AS gobuild
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o server ./main.go

# Final runtime image
FROM alpine:3.22
WORKDIR /app
RUN apk add --no-cache ca-certificates
# copy Backend binary
COPY --from=gobuild /app/server /app/server
# copy Frontend dist
COPY --from=webbuild /app/web/dist /app/web/dist
ENV ADDR=:8080
EXPOSE 8080
CMD ["/app/server"]
