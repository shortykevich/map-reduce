.PHONY: create/mr-tmp clean/out clean/plugins build/plugin run/seq run/mrcoordinator run/mrworker build/mrcoordinator build/mrworker build/mr

PDIR := "./mrapps"
PLUGINS := crash.go early_exit.go indexer.go jobcount.go mtiming.go nocrash.go rtiming.go wc.go

create/mr-tmp:
	@mkdir -p ./main/mr-tmp

clean/out:
	@rm -f ./main/mr-tmp/*

clean/plugins:
	@rm -f ./mrapps/*.so

build/plugins: | clean/plugins
	@for plugin in ${PLUGINS}; do \
	go build -race -o ${PDIR} -buildmode=plugin $(PDIR)/$$plugin; \
	done

build/seq: | clean/out
	@go build -race -o ./main/ ./main/mrsequential.go

build/mrcoordinator:
	@rm -f mrcoordinator
	@go build -race -o ./main/ ./main/mrcoordinator.go

build/mrworker:
	@rm -f mrworker
	@go build -race -o ./main/ ./main/mrworker.go

build/mr: create/mr-tmp build/plugins build/mrcoordinator build/mrworker

run/seq: | clean/out
	@cd main/mr-tmp && ../mrsequential ../../mrapps/wc.so ../fixtures/pg*.txt

run/mrcoordinator:
	@cd main/mr-tmp && ../mrcoordinator ../fixtures/pg*.txt

run/mrworker:
	@cd main/mr-tmp && ../mrworker ../../mrapps/wc.so

run/tests: | clean/out
	@cd main/ && ./test-mr.sh
