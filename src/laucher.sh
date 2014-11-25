#!/bin/bash
if [ "$#" -ne 2 ]; then
    echo Usage : $0 ' <Config File Number (1~8)> <number of banks>'
    exit
fi

pkill server
pkill master
pkill client
rm ../logs/*
if [ $1 -lt 10 ]; then
    N="0$1"
    echo $N
else
    N="$1"
fi
for i in `seq 0 $[$2-1]`; do
    for j in `seq $[(4+$i)*1000+1] $[(4+$i)*1000+4]`; do
        go run server/server.go $j config$N.json &
    done
done
gnome-terminal -e "go run master/master.go config$N.json > m.txt"
#for i in `seq 0 $[$2-1]`; do
#    PORT=$[(4+$i)*1000+999]
#    echo $PORT
#    gnome-terminal -e "go run client/client.go $PORT config$N.json"
#done
gnome-terminal -e "go run client/client.go 4999 config$N.json"
echo servers launched!
