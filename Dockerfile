# Build
FROM golang:1.17 AS builder
WORKDIR /src
COPY go.* main.go /src/
COPY ./mod /src/mod
RUN go build -o kuta .

# Ubuntu + Kuta
FROM ubuntu:20.04

# Install sudo
RUN apt-get update -q && apt-get install -qy sudo

# Allow all users to run sudo.
RUN echo "ALL ALL=NOPASSWD: ALL" >> /etc/sudoers

# Install entrypoint
COPY --from=builder /src/kuta /kuta
RUN chown 0:0 /kuta && chmod +s /kuta

ARG USERNAME=ubuntu
ARG USER_UID=1300
ARG USER_GID=$USER_UID

# Create ubuntu user
RUN groupadd --gid "$USER_GID" "$USERNAME"

RUN useradd \
  --uid "$USER_UID" \
  --gid "$USER_GID" \
  --create-home \
  "$USERNAME"

USER $USER_UID:$USER_GID
WORKDIR /home
ENV USER=ubuntu
ENV KUTA_DEBUG=1

ENTRYPOINT ["/kuta"]
