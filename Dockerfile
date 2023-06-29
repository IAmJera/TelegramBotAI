FROM golang:1.20.4
RUN mkdir /app
ADD . /app
WORKDIR /app
RUN apt-get update && apt-get install -y ffmpeg
RUN go build -o main ./app/bot.go
ENTRYPOINT ["/app/main"]