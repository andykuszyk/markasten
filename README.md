# markasten
> A Zettelkasten tool for Markdown files.

## Installation
`markasten` is a command line tool which can either be installed locally, or used via a Docker image.

To install locally using the Go toolchain:
```sh
go install github.com/andykuszyk/markasten/cmd/markasten@master
markasten --help
```

Or, to use via a Docker image:
```sh
docker run andykuszyk/markasten:latest markasten --help
```

## Usage
### Generate an index of tags from some files
The `tags` command is used to generate an index of files based on tags present in a header in each file. Headers are YAML formatted, and are expected to appear as frontmatter/metadata at the top of the file, enclosed in `---`. Tags are expected in the `tags` key.

An example of a header is as follows:

```markdown
---
tags:
- tag-one
- tag-two
---
```

The `tags` command can be invoked using the CLI:
```sh
markasten tags -i <path-to-input-files> -o <path-to-index-file> -t <index-title>
```

Or via the Docker image:
```sh
docker run -v "$(pwd)":/input -v "$(pwd)":/output andykuszyk/markasten:latest markasten tags --capitalize -i /input -o /output/README.md 
```

It supports the following flags:
```sh
$ markasten tags --help
Usage:
  markasten tags [flags]

Flags:
      --capitalize      If set, tag names in the generated index will have their first character capitalized.
      --debug           If set, debug logging will be enabled
  -h, --help            help for tags
  -i, --input string    The location of the input files
  -o, --output string   The location of the output files
  -t, --title string    The title of the generated index file (default "Index")
      --wiki-links      If set, links will be generated for a wiki with file extensions excluded
```

It can also be invoked using the GitHub Action in this repo:
```yaml
name: docs
on: workflow_dispatch
jobs:
  tags:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: andykuszyk/markasten@master
        with:
          command: "tags"
          input: "docs/"
          output: "docs/INDEX.md"
          additionalArgs: "--capitalize --wiki-links"
      - run: cat docs/INDEX.md
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
