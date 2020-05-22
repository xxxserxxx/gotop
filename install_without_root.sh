#!/bin/sh

VERSION='1.14.2'
OS='linux'
ARCH='amd64'

GO_NAME=go${VERSION}.${OS}-${ARCH}

mkdir -p $HOME/pkg
cd $HOME/pkg

curl https://dl.google.com/go/${GO_NAME}.tar.gz --output ./${GO_NAME}.tar.gz

tar -vxzf ${GO_NAME}.tar.gz
rm ${GO_NAME}.tar.gz

PATH=$HOME/pkg/go/bin:$PATH

go env -w GOPATH=$HOME/pkg # otherwise go would create a directory in $HOME

rm -rf ./gotop
git clone https://github.com/xxxserxxx/gotop.git
cd ./gotop
go build -o gotop ./cmd/gotop

go clean -modcache # otherwise $HOME/pkg/pkg would need sudo permissions to remove

ln -sf $HOME/pkg/gotop/gotop $HOME/bin/gotop

rm -rf $HOME/pkg/go
rm -rf $HOME/pkg/pkg # installed as part of go
