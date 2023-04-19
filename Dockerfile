FROM golang@sha256:31a2f928187818ac6f965640c18bb9c6460a69cbca7ca6456a50d720edf0928c AS build
WORKDIR markasten
COPY ./ ./
RUN mkdir /input  /output
RUN go build -o markasten ./cmd/markasten/main.go && mv markasten /usr/local/bin
FROM gcr.io/distroless/base-debian11@sha256:e711a716d8b7fe9c4f7bbf1477e8e6b451619fcae0bc94fdf6109d490bf6cea0
COPY --from=build /usr/local/bin/markasten /usr/local/bin/markasten
COPY --from=build /input /input
COPY --from=build /output /output
CMD markasten --help
