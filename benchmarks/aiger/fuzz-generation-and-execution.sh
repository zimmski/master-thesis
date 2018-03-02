#!/bin/sh

while true
do
	rm -rf testset-* || true
	go run generate-testset.go || exit
	go run execute-testset.go || exit
done
