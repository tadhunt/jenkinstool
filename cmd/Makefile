all:
	go mod tidy
	go build -o jenkins

install: all
	cp jenkins ${HOME}/bin
