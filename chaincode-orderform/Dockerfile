FROM golang:1.15-alpine3.12

WORKDIR /go/src

# Download Go modules
COPY ./src/go.mod .
COPY ./src/go.sum .
RUN go mod download

COPY ./src/*.go ./

# Build
RUN go install

# For change timezone
RUN apk update && apk add tzdata

# EXPOSE 9995

CMD ["chaincode-orderform"]
