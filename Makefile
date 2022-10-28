static:
	CC=musl-gcc go build \
		--ldflags '-linkmode external -extldflags "-static"' \
		-tags timetzdata \
		-o jeff main.go
