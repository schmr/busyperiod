gitdate := $(shell git log -1 --format=\"%ad\" --date=short)
gitrevision := $(shell git log -1 --format=\"%h\")

.PHONY: run test lint race

busyp: cmd/busyp/main.go taskset/taskset.go | test
	go build -o $@ -ldflags "-X main.revisiondate=${gitdate} -X main.revision=${gitrevision}" $<

run: cmd/busyp/main.go taskset/taskset.go | test
	go run -ldflags "-X main.revisiondate=${gitdate} -X main.revision=${gitrevision}" $<

test: taskset/taskset.go taskset/printing_test.go taskset/calculation_test.go | lint
	cd taskset && go test

lint: taskset/taskset.go taskset/printing_test.go taskset/calculation_test.go
	goimports -w $^
	golint -set_exit_status

race: cmd/busyp/main.go taskset/taskset.go | test
	go run -race cmd/busyp/main.go

