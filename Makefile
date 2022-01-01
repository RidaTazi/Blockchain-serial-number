all: setup build

build: 
	go build -o ./bin/bcg *.go

clean: 
	rm -rf ./bin/bcg

setup: 
	go get
	go install
	go get github.com/qihengchen/blockchain_go/...
	go get github.com/boltdb/bolt/...
	go get github.com/fatih/color/...
	go get -u golang.org/x/crypto/...
	# used 'bcg' in scripts, so renaming the directory
	cp -r  ../github.com/qihengchen/blockchain_go ../github.com/qihengchen/bcg
	rm -R ../github.com/qihengchen/blockchain_go
	go install github.com/qihengchen/bcg

test: 
	go test
