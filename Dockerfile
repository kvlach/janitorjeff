FROM golang:1.22-alpine AS build-env
RUN apk add -U --force-refresh --no-cache --purge --clean-protected -l -u gcc musl-dev
WORKDIR /go/src/github.com/janitorjeff/jeff-bot
COPY . .
RUN go build -o /go/bin/jeff

FROM alpine:latest
RUN apk add --no-cache tzdata ffmpeg yt-dlp
WORKDIR /app
COPY --from=build-env /go/bin/jeff ./
COPY schema.sql .
ENTRYPOINT ["./jeff"]
