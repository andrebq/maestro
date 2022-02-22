.PHONY: test benchmar tidy watch default


default: test

test:
	go test ./...

benchmark:
	go test -test.bench 'Benchmark*' ./...

tidy:
	go mod tidy
	go fmt ./...

watch:
	modd -f modd.conf
