#版本号

SRCS = $(shell git ls-files '*.go')


## Format the Code
gofmt:
	gofmt -s -l -w $(SRCS)
