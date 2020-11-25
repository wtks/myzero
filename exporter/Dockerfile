FROM golang:1.15-alpine AS build
RUN apk add --update --no-cache git
WORKDIR /src
COPY ./go.* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /exporter

FROM scratch
WORKDIR /
EXPOSE 8080
COPY --from=build /exporter /exporter
ENTRYPOINT ["/exporter"]