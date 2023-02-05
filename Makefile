build:
	go build -ldflags="-w -s" twmd.go

windows-gui-action:
	GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc CXX=x86_64-w64-mingw32-g++  go  build -o twmd-GUI.exe gui.go
	cp twmd-GUI.exe build-artifacts*/.

windows-gui:
	GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc CXX=x86_64-w64-mingw32-g++  go  build -o twmd-GUI.exe gui.go

linux-gui:
	GOOS=linux go build -o twmd-GUI gui.go

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
