test:
	go test -coverprofile=output.txt ./client ./pandatypes
	gcov2lcov -infile output.txt -outfile lcov.info
	rm output.txt

test-verbose:
	go test -v -coverprofile=output.txt ./client ./pandatypes
	gcov2lcov -infile output.txt -outfile lcov.info
	rm output.txt

# Real-Postgres tests; needs Docker.
test-db:
	go test -race -tags=dbtest ./client/...

lint:
	golangci-lint run ./...

fmt:
	goimports -w -local github.com/feimaomiao/stalka ./

fmt-check:
	golangci-lint run --no-fix ./...

pre-commit-install:
	pre-commit install

pre-commit-run:
	pre-commit run --all-files

.PHONY: test test-verbose test-db lint fmt fmt-check pre-commit-install pre-commit-run
