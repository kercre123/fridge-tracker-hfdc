#!/bin/bash

if [[ -f source.env ]]; then
	source source.env
else
	echo "No source file"
fi
if [[ -f main ]]; then
	echo "Starting compiled program"
	./main
else
	echo "Starting via go run"
	go run main.go
fi
