
PROGS=\
	bin/snes-img

all: bin/ $(PROGS)
bin/:
	-mkdir bin

clean:
	-rm $(PROGS)

bin/%: cmd/%.go
	go build -o $@ $^
