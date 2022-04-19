FROM golang:1.17.9-alpine3.15
EXPOSE 8080
WORKDIR /opt
ADD . /opt/
RUN go get

CMD "go" "run" "main.go"
