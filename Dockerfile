# Build
FROM golang:1.17 AS builder
WORKDIR /src
COPY go.* main.go /src/
COPY ./mod /src/mod
RUN go build -o kuta .

# Ubuntu + Kuta
FROM ubuntu:20.04

# Install entrypoint
COPY --from=builder /src/kuta /kuta
RUN chown 0:0 /kuta && chmod +s /kuta

# Create ubuntu user
RUN adduser ubuntu
USER ubuntu
WORKDIR /home/ubuntu
ENV USER=ubuntu

ENTRYPOINT ["/kuta"]
