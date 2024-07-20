FROM golang:1.22-alpine as builder

ENV CGO_ENABLED 0
ENV GOOS linux
ENV GOPROXY https://goproxy.cn,direct

WORKDIR /build

# Copy and download dependency using go mod.
ADD ../go.mod .
ADD ../go.sum .
RUN go mod download

# Copy the code into the container.
COPY . .

# Set necessary environment variables needed for our image
# and build the API server.
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN go build -ldflags="-s -w" -o webhook-server ./k8s-dac/webhook-server

FROM alpine

RUN apk update --no-cache && apk add --no-cache ca-certificates tzdata
ENV TZ Asia/Shanghai

WORKDIR /app

# Copy binary and config files from /build
# to root folder of scratch container.
COPY --from=builder ["/build/webhook-server", "/"]

CMD [ "/webhook-server" ]