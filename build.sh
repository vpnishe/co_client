#!/bin/bash
GOOS=windows GOARCH=amd64 go build -o bin/co_cc-amd64-windows.exe
GOOS=windows GOARCH=386 go build -o bin/co_cc-386-windows.exe
GOOS=windows GOARCH=arm go build -o bin/co_cc-arm-windows.exe


GOOS=linux GOARCH=amd64 go build -o bin/co_cc-amd64-linux
GOOS=linux GOARCH=386 go build -o bin/co_cc-386-linux
GOOS=linux GOARCH=arm go build -o bin/co_cc-arm-linux
GOOS=linux GOARCH=arm64 go build -o bin/co_cc-arm64-linux
GOOS=linux GOARCH=mips go build -o bin/co_cc-mips-linux
GOOS=linux GOARCH=mips64 go build -o bin/co_cc-mips64-linux

GOOS=darwin GOARCH=amd64 go build -o bin/co_cc-amd64-darwin
GOOS=darwin GOARCH=arm64 go build -o bin/co_cc-arm64-darwin

GOOS=android GOARCH=amd64 go build -o bin/co_cc-amd64-android
GOOS=android GOARCH=386 go build -o bin/co_cc-386-android
GOOS=android GOARCH=arm64 go build -o bin/co_cc-arm64-android

for filename in bin/*; do
    [ -e "$filename" ] || continue
    zip "$filename.zip" "$filename"
    rm "$filename" 
done

