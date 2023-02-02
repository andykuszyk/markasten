FROM golang:1.19
WORKDIR markasten
COPY ./ ./
RUN go build -o markasten ./cmd/markasten/main.go && mv markasten /usr/local/bin
RUN mkdir /input  /output
CMD markasten --help
