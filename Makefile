.PHONY: clean-out build/plugin run/seq run/mrcoordinator run/mrworker build/mrcoordinator build/mrworker build/mr

clean-out:
	@rm -f mr-out*

build/plugin:
	@rm -f wc.so
	@go build -o wc.so -buildmode=plugin ./mrapps/wc.go

build/mrcoordinator:
	@rm -f coord
	@go build -o coord ./mrcoordinator/mrcoordinator.go

build/mrworker:
	@rm -f worker
	@go build -o worker ./mrworker/mrworker.go

build/mr: build/plugin build/mrcoordinator build/mrworker

run/seq: | clean-out
	@go run mrsequential.go wc.so pg*.txt

run/mrcoordinator:
	@go run coord pg-*.txt

run/mrworker:
	@go run worker wc.so
