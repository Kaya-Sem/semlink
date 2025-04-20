
.PHONY: test coverage coverage-html clean

test:
	go test ./... -cover -coverprofile=coverage.out

coverage: test
	go tool cover -func=coverage.out

coverage-html: test
	go tool cover -html=coverage.out -o coverage.html

clean:
	rm -f coverage.out coverage.html
