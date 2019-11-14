test:
	go test -short ./...

bench:
  go test -run=XXX -bench=^Benchmark -benchmem -memprofile prof.mem -cpuprofile prof.cpu -benchtime 10s -v
