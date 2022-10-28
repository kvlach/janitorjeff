static:
	CC=musl-gcc go build --ldflags '-linkmode external -extldflags "-static"' -o jeff main.go
