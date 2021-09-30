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

# Create ubuntu user
ARG USERNAME=ubuntu
ARG USER_UID=1300
ARG USER_GID=$USER_UID
RUN groupadd --gid "$USER_GID" "$USERNAME"
RUN useradd \
  --uid "$USER_UID" \
  --gid "$USER_GID" \
  --create-home \
  "$USERNAME"

# Switch to user
USER $USER_UID:$USER_GID
# Needed by kuta to pick the user to mutate.
ENV USER=ubuntu
# Enable debug logs
ENV KUTA_DEBUG=1
# Test that kuta keeps the same PWD
WORKDIR /home

# Go!
ENTRYPOINT ["/kuta"]
