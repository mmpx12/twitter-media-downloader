build:
	go build -ldflags="-w -s"

install:
	mv twitter-media-downloader /usr/bin/twmd

termux-install:
	mv twitter-media-downloader /data/data/com.termux/files/usr/bin/twmd

all: build install

termux-all: build termux-install

clean:
	rm -f twitter-media-downloader /usr/bin/twmd

termux-clean:
	rm -f twitter-media-downloader /data/data/com.termux/files/usr/bin/twmd
