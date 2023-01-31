# markasten
> A Zettelkasten tool for Markdown files.

## Installation
```sh
go install github.com/andykuszyk/markasten/cmd/markasten@master
```

## Usage
### Generate an index of tags from some files
```sh
markasten tags -i <path-to-input-files> -o <path-to-index-file> -t <index-title>
```

### Find backlinks amongst files (TODO)
```sh
markasten backlinks find -i <path-to-input-files> -o <path-to-output-files>
```

### Append backlinks idempotently to existing files (TODO)
```sh
markasten backlinks append -i <path-to-backlink-files> -o <path-to-target-files>
```

## Development
1. Clone this repo.
2. Run `go test ./...`
