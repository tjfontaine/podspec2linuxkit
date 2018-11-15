all: podspec2linuxkit

podspec2linuxkit: cmd/podspec2linuxkit.go pkg/linuxkit/config.go
	go build -o ./podspec2linuxkit cmd/*

clean:
	rm ./podspec2linuxkit
