# backend/Dockerfile
FROM golang:1.24-alpine

WORKDIR /app

# 1) Pull dependencies in a separate cache
COPY go.mod go.sum ./
RUN go mod download

# 2) Copy code
COPY . .

# 3) Build the cmd package (where main.go is located)
# CGO disabled → static binary, works great for Echo + mysql driver
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd

EXPOSE 8080
CMD ["./server"]
