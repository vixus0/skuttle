from golang:1.16-alpine as build
workdir /opt/build

copy go.mod go.sum /opt/build/
run go mod download

copy cmd /opt/build/cmd
copy internal /opt/build/internal
run export CGO_ENABLED=0 \
 && go build ./cmd/skuttle

from scratch as run
copy --from=build /opt/build/skuttle /skuttle

entrypoint ["/skuttle"]
