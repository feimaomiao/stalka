test:
	go test -coverprofile=output.txt ./client ./pandatypes
	gcov2lcov -infile output.txt -outfile lcov.info
	rm output.txt
test-verbose:
	go test -v -coverprofile=output.txt ./client ./pandatypes
	gcov2lcov -infile output.txt -outfile lcov.info
	rm output.txt