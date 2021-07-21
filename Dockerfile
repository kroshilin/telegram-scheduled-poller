FROM golang:alpine
RUN apk add git

WORKDIR /go/src/app
COPY . .

RUN go env -w GO111MODULE=off
RUN go get -d -v ./...
RUN go install -v ./...

CMD ["app"]
