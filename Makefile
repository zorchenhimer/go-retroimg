
PROGS=\
	  bin/img2chr \
	  bin/chr2img \
	  bin/extractchr \

all: bin/ $(PROGS)
bin/:
	-mkdir bin

clean:
	-rm $(PROGS)

bin/%: cmd/%.go *.go palette/*.go
	go build -o $@ $<
