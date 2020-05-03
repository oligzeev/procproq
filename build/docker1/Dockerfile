# sudo docker run --name jgr -p5775:5775/udp -p6831:6831/udp -p16686:16686 -d jaegertracing/all-in-one:latest
# sudo docker run --name pg -p5432:5432 -e POSTGRES_PASSWORD=pg -e POSTGRES_USER=pg -e POSTGRES_DB=pg -d postgres
# sudo docker build -t pp-gin -f build/docker1/Dockerfile .
# sudo docker run --rm --link pg:pg -e PP_DB_HOST=pg pp-gin

FROM golang:latest

RUN mkdir /app
COPY . /app
WORKDIR /app
RUN go build -o main /app/cmd/pp-gin/

CMD ["/app/main"]