FROM golang:alpine AS builder
RUN apk update && apk add --no-cache git
WORKDIR $GOPATH/src/mypackage/myapp/
COPY . .
RUN go get -d -v
RUN go build -mod=vendor -o /go/bin/client

FROM scratch
COPY --from=builder /go/bin/client /go/bin/client
ENTRYPOINT ["/go/bin/client"]
