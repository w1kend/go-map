test:
	go test .

bench:
	go test . -run=$^ -bench . -benchmem