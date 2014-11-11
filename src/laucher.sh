#!/bin/bash
#if [ "$#" -ne 2 ]; then
#    echo Usage : $0 ' <min server id> <max server id>'
#    exit
#fi

pkill server
rm ../logs/*
#xterm go run master/master.go
gnome-terminal -e "go run master/master.go config06.json"
#for i in `seq $1 $2`; do
for i in `seq 4001 4004`; do
#    server $i &
    go run server/server.go $i config06.json &
done
gnome-terminal -e "go run client/client.go 4999 config06.json"
echo servers launched!
