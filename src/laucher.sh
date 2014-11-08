#!/bin/bash
#if [ "$#" -ne 2 ]; then
#    echo Usage : $0 ' <min server id> <max server id>'
#    exit
#fi

pkill server
#xterm go run master/master.go
gnome-terminal -e "go run master/master.go"
#for i in `seq $1 $2`; do
for i in `seq 4001 4003`; do
#    server $i &
    go run server/server.go $i &
done
gnome-terminal -e "go run client/client.go 4999"
echo servers launched!
