build:
	go build -ldflags="-w -s"

install:
	mv twmd /usr/bin/twmd

termux-install:
	mv twmd /data/data/com.termux/files/usr/bin/twmd

all: build install

termux-all: build termux-install

clean:
	rm -f twmd /usr/bin/twmd

termux-clean:
	rm -f twmd /data/data/com.termux/files/usr/bin/twmd
