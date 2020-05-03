# Tested on docker 19.03
# sudo docker build -t pp-gin -f build/docker2/Dockerfile .
# sudo docker run --rm --link pg:pg -e PP_DB_HOST=pg -e PP_REST_HOST=0.0.0.0 -p9090:8080 pp-gin

FROM golang:latest AS builder
RUN mkdir /app
COPY . /app
WORKDIR /app
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main /app/cmd/pp-gin/

FROM alpine:latest
RUN mkdir -p /app/config
COPY --from=builder /app/config/pp-gin.yaml /app/config
COPY --from=builder /app/main /app
WORKDIR /app
CMD ["/app/main"]