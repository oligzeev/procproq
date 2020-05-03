# Tested on docker 19.03
# sudo docker build --build-arg GIT_REPO=https://github.com/oligzeev/procproq.git -t pp-gin -f build/docker3/Dockerfile .
# sudo docker run --rm --link pg:pg -e PP_DB_HOST=pg -e PP_REST_HOST=0.0.0.0 -e PP_REST_PORT=9090 -p9090:9090 pp-gin
# http://localhost:9090/swagger/index.html

FROM golang:latest AS builder
ARG GIT_REPO
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64
RUN mkdir /app
WORKDIR /app
RUN git clone -v $GIT_REPO /app
RUN go build -a -installsuffix cgo -o main /app/cmd/pp-gin/

FROM alpine:latest
RUN mkdir -p /app/config
COPY --from=builder /app/config/pp-gin.yaml /app/config
COPY --from=builder /app/main /app
WORKDIR /app
EXPOSE $PP_REST_PORT
CMD ["/app/main"]