# 阶段1: 构建前端
FROM node:18-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package.json frontend/package-lock.json* ./
RUN npm install
COPY frontend/ ./
RUN npm run build

# 阶段2: 构建后端
FROM golang:1.21-alpine AS backend-builder
ARG COMMIT_ID=unknown
ARG BUILD_TIME=unknown
WORKDIR /app/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-X main.commitID=${COMMIT_ID} -X main.buildTime=${BUILD_TIME}" \
    -o server .

# 阶段3: 运行阶段
FROM alpine:3.18
RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Shanghai

WORKDIR /app

COPY --from=backend-builder /app/backend/server .
COPY --from=frontend-builder /app/frontend/dist ./frontend/dist

RUN mkdir -p uploads saved_html

EXPOSE 8080

CMD ["./server"]
