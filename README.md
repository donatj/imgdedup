Image De-duplicator
===================

[![Go Report Card](https://goreportcard.com/badge/donatj/imgdedup)](https://goreportcard.com/report/donatj/imgdedup)
[![Build Status](https://travis-ci.org/donatj/imgdedup.svg?branch=master)](https://travis-ci.org/donatj/imgdedup)

Simple image de-duplication in Go-lang

	usage: imgdedup [options] [<directories>/files]
	  -diff="": Command to pass dupe images to eg: cmd $left $right
	  -subdivisions=10: Slices per axis
	  -tolerance=100: Color delta tolerance, higher = more tolerant

## Features

Detects duplications despite changes in

- size
- quality
- aspect ratio

Compare images in your prefered difftool

## Download

### Binaries
	
See: [Releases](https://github.com/donatj/imgdedup/releases).

### Compile

	$ go get github.com/donatj/imgdedup
