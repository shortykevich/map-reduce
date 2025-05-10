.PHONY: clean-out run/seq run/mrcoordinator run/mrworker

clean-out:
	@rm -f mr-out*

run/seq: | clean-out
	@go run mrsequential.go wc.so pg*.txt

run/mrcoordinator: | clean-out
	@go run ./mrcoordinator/mrcoordinator.go ../pg-*.txt

run/mrworker: | clean-out
	@go run ./mrworker/mrworker.go ../pg-*.txt
