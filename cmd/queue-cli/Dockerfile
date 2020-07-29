FROM alpine:latest
RUN apk add ca-certificates

RUN mkdir /app
WORKDIR /app
ADD queue-cli ./

CMD ["./queue-cli"]