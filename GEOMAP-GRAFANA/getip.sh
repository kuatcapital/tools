#!/bin/bash

echo 
ls -1 *.toml
directory="/root/namada-testnets/namada-public-testnet-15"


> xoutput.txt


for filename in "$directory"/*.toml; do
    if [ -f "$filename" ]; then

        file_name=$(basename "$filename" .toml)


        ip_addresses=$(grep -oE "\b([0-9]{1,3}\.){3}[0-9]{1,3}\b" "$fil
ename")

        echo "Name: $file_name, IP: $ip_addresses" >> xoutput.txt
    fi
done


