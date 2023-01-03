#!/bin/bash

source source.env
if [[ -f main ]]; then
	./main
else
	/usr/local/go/bin/go run main.go
fi
