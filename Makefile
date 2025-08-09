test:
	go test -coverprofile=output.txt ./client ./pandatypes -v
	gcov2lcov -infile output.txt -outfile lcov.info
	rm output.txt