FROM golang:latest as builder

LABEL maintainer="Emre Savcı <freelancemuhendis@gmail.com>"

WORKDIR $GOPATH/src/mypackage/consumerapp/
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w' -o consumer .

FROM alpine
RUN adduser -S -D -H -h /app appuser
USER appuser
COPY --from=builder /go/src/mypackage/consumerapp /app
WORKDIR /app
CMD ["./consumer"]