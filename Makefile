INSTALL_PATH=/opt/svt/bin/fctl

all: clean cleanfission build install

clean:
	mkdir -p target 
	rm -rf target/*
	rm -f $(INSTALL_PATH)

cleanfission:
	fission fn delete --name "monitor" --ignorenotfound
	fission env delete --name "invoker" --ignorenotfound
	-fission httptrigger delete --name "monitor-load-root-config" --ignorenotfound

build:
	cd fctl && go build -o ../target/fctl .

install:
	rm -f $(INSTALL_PATH)
	cp target/fctl $(INSTALL_PATH)
