
PROGS=\
	  bin/chr2img \
	  bin/extractchr \
	  bin/img2chr \
	  bin/img2screen \

all: bin/ $(PROGS)
bin/:
	-mkdir bin

clean:
	-rm $(PROGS)

bin/%: cmd/%.go *.go palette/*.go
	go build -o $@ $<
