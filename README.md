Image De-duplicator
===================

Simple image de-duplication in Go-lang

	usage: imgdedup [options] [<directories>/files]
	  -subdivisions=10: Slices per axis
	  -tolerance=100: Color delta tolerance, higher = more tolerant

## Features

Detects duplications despite changes in

- size
- quality
- aspect ratio

## Download

	go get github.com/donatj/imgdedup