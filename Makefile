all:
	go mod tidy
	go test -v ./...
	cd cmd && make

install: all
	cd cmd && make install
