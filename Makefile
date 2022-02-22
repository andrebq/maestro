.PHONY: test tidy watch default


default: test

test:
	go test ./...

tidy:
	go mod tidy
	go fmt ./...

watch:
	modd -f modd.conf
