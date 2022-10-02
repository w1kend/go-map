test:
	go test .

bench:
	go test . -run=^$$ -bench . -benchmem

fuzz:
	go test -run=^$$ -fuzz FuzzMap