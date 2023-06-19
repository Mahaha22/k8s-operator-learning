FROM golang:1.19 as builder
WORKDIR /app
COPY . .
RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN CGO_ENABLED=0 go build -o ingress-manager main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/ingress-manager .
CMD ["./ingress-manager"]