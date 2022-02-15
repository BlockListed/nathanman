FROM golang:latest as builder
WORKDIR /usr/src/nathanman

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN CGO_ENABLED=0 go build -v -o /usr/local/bin/nathanman

FROM s6on/alpine
COPY --from=builder /usr/local/bin/nathanman /usr/local/bin/nathanman
COPY rootfs/ /