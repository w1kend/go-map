test:
	go test .

bench:
	go test . -run=^$$ -bench . -benchmem

bench-overflow:
	go test . -run=^$$ -bench ^BenchmarkPutWithOverflow$$ -benchmem

fuzz:
	go test -run=^$$ -fuzz FuzzMap