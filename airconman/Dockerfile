FROM golang:1.15.0-alpine AS build
RUN apk add --update --no-cache git
WORKDIR /src
COPY ./go.* ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 go build -o /airconman -ldflags "-s -w"

FROM alpine
COPY --from=build /airconman /usr/local/bin/
ENTRYPOINT ["airconman"]
