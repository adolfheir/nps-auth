
go install github.com/go-delve/delve/cmd/dlv@latest

###
docker run  -w $(pwd) -it  -v "$(pwd):$(pwd)" -v $(pwd)/.cache/gocache:/gocache -p 7998:7998 -p 7999:7999 ihouqi-docker.pkg.coding.net/polaris/cyh/go_devel:1.20.2 bash

dlv debug  --listen=:7998  --api-version=2 --accept-multiclient -- app
<!-- docker pull golang:1.23rc1-bullseye -->
dlv debug --headless --listen=:7998 --api-version=2 --accept-multiclient -- app
