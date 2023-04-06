FROM golang:1.19 AS build
WORKDIR markasten
COPY ./ ./
RUN mkdir /input  /output
RUN go build -o markasten ./cmd/markasten/main.go && mv markasten /usr/local/bin
FROM gcr.io/distroless/base-debian11
COPY --from=build /usr/local/bin/markasten /usr/local/bin/markasten
COPY --from=build /input /input
COPY --from=build /output /output
CMD markasten --help
