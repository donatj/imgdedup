Image De-duplicator
===================

[![Go Report Card](https://goreportcard.com/badge/donatj/imgdedup)](https://goreportcard.com/report/donatj/imgdedup)
[![CI](https://github.com/donatj/imgdedup/actions/workflows/ci.yml/badge.svg)](https://github.com/donatj/imgdedup/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/donatj/imgdedup.svg)](https://pkg.go.dev/github.com/donatj/imgdedup)

Simple image de-duplication

```
Usage of imgdedup [options] [<directories>/files]:
  -cache-dir string
         (default "$HOME/.cache/imgdedup/cacheDb")
  -diff string
        Command to pass dupe images to eg: cmd $left $right
  -format string
        Output format - options: default json (default "default")
  -subdivisions uint
        Slices per axis (default 10)
  -tolerance uint
        Color delta tolerance, higher = more tolerant (default 100)
```

## Features

Detects duplications despite changes in

- size
- quality
- aspect ratio

Flags to compare images in your prefered difftool

## Download

### Binaries
	
See: [Releases](https://github.com/donatj/imgdedup/releases).

### Compile

	$ go install github.com/donatj/imgdedup@latest
