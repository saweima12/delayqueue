cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	firefox ./coverage.html

clean:
	rm ./coverage*
