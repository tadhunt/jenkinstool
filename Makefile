all:
	go mod tidy
	go test ./...
	cd cmd && make

install: all
	cd cmd && make install
