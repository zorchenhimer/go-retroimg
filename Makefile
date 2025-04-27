
PROGS=\
	  bin/img2chr \
	  bin/chr2img

all: bin/ $(PROGS)
bin/:
	-mkdir bin

clean:
	-rm $(PROGS)

bin/%: cmd/%.go image/*.go
	go build -o $@ $<
