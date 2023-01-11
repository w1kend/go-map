test:
	go test .

bench:
	go test . -run=^$$ -bench . -benchmem

bench-overflow:
	go test . -run=^$$ -bench ^BenchmarkPutWithOverflow$$ -benchmem

bench-overflow-profile:
	go test . -run=^$$ -bench ^BenchmarkPutWithOverflow$$ -benchmem -cpuprofile cpu.out && \
	go tool pprof -http :8080 cpu.out

bench-overflow-memprofile:
	go test . -run=^$$ -bench ^BenchmarkPutWithOverflow$$ -benchmem -memprofile mem.out && \
	go tool pprof -http :8080 mem.out

fuzz:
	go test -run=^$$ -fuzz FuzzMap