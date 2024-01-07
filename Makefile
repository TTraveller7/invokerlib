INSTALL_PATH=/opt/svt/bin/fctl

all: clean build install

clean:
	mkdir -p target 
	rm -rf target/*
	rm -f $(INSTALL_PATH)

build:
	cd fctl && go build -o ../target/fctl .

install: 
	rm -f $(INSTALL_PATH)
	cp target/fctl $(INSTALL_PATH)
