.PHONY: clean-out build/plugin run/seq run/mrcoordinator run/mrworker build/mrcoordinator build/mrworker build/mr

clean-out:
	@rm -f mr-out*

build/plugin:
	@rm -f wc.so
	@go build -o wc.so -buildmode=plugin ./mrapps/wc.go

build/seq: | clean-out
	@go build mrsequential.go

build/mrcoordinator:
	@rm -f coord
	@go build -race -o coord ./mrcoordinator/mrcoordinator.go

build/mrworker:
	@rm -f worker
	@go build -race -o worker ./mrworker/mrworker.go

# build/plugin
build/mr: build/mrcoordinator build/mrworker

run/seq: | clean-out
	@./mrsequential wc.so pg*.txt

run/mrcoordinator:
	@./coord pg-*.txt

run/mrworker:
	@./worker
