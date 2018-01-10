.PHONY: build run

.DEFAULT_GOAL := run

build:
	go build -o delaystat

run: build
	sudo ./delaystat
