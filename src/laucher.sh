#!/bin/bash
if [ "$#" -ne 1 ]; then
    echo Usage : $0 ' <Config File Number (1~8)>'
    exit
fi

pkill server
pkill master
pkill client
rm ../logs/*
gnome-terminal -e "go run master/master.go config0$1.json"
for i in `seq 4001 4004`; do
    go run server/server.go $i config0$1.json &
done
for i in `seq 5001 5004`; do
    go run server/server.go $i config0$1.json &
done
gnome-terminal -e "go run client/client.go 4999 config0$1.json"
gnome-terminal -e "go run client/client.go 5999 config0$1.json"
echo servers launched!
