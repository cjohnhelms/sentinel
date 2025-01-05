#!/bin/bash

# apt
sudo apt-update > /dev/null
sudo apt install -y wget > /dev/null

# install golang
wget -c https://golang.org/dl/go1.23.4.linux-arm64.tar.gz
rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.23.4.linux-arm64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# check
go version
if [[ $? -eq 0 ]]; then
    echo "INSTALL OK"
else
    echo "INSTALL FAILED"
fi

