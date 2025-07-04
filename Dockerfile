# syntax=docker/dockerfile:1

FROM golang:1.22-alpine AS build
WORKDIR /app
COPY . .
RUN go build -o /bin/git-mfpr ./cmd/git-mfpr

FROM alpine:latest
COPY --from=build /bin/git-mfpr /usr/local/bin/git-mfpr
ENTRYPOINT ["git-mfpr"] 