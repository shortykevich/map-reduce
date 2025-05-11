.PHONY: clean-out build/plugin run/seq run/mrcoordinator run/mrworker

clean-out:
	@rm -f mr-out*

build/plugin:
	@rm -f ./mrworker/wc.so
	@go build -o mrworker/wc.so -buildmode=plugin ./mrapps/wc.go

run/seq: | clean-out
	@go run mrsequential.go ./mrworker/wc.so pg*.txt

run/mrcoordinator:
	@go run ./mrcoordinator/mrcoordinator.go ../pg-*.txt

run/mrworker:
	@go run ./mrworker/mrworker.go ./mrworker/wc.so
