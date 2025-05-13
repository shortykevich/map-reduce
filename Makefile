.PHONY: clean/out clean/plugins build/plugin run/seq run/mrcoordinator run/mrworker build/mrcoordinator build/mrworker build/mr

PDIR := "./mrapps"
PLUGINS := crash.go early_exit.go indexer.go jobcount.go mtiming.go nocrash.go rtiming.go wc.go

clean/out:
	@rm -f mr-out*

clean/plugins:
	@rm -f ./mrapps/*.so

build/plugins: | clean/plugins
	@for plugin in ${PLUGINS}; do \
	go build -race -o ${PDIR} -buildmode=plugin $(PDIR)/$$plugin; \
	done

build/seq: | clean/out
	@go build mrsequential.go

build/mrcoordinator:
	@rm -f coord
	@go build -race -o coord ./mrcoordinator/mrcoordinator.go

build/mrworker:
	@rm -f worker
	@go build -race -o worker ./mrworker/mrworker.go

build/mr: build/plugins build/mrcoordinator build/mrworker

run/seq: | clean/out
	@./mrsequential wc.so pg*.txt

run/mrcoordinator:
	@./coord pg-*.txt

run/mrworker:
	@./worker
