INSTALL_PATH=/opt/svt/bin/fctl

all: cleanall applyk8s build install

cleanall: clean cleanfission cleank8s

clean:
	mkdir -p target 
	rm -rf target/*
	rm -f $(INSTALL_PATH)

cleanfission:
	fission fn delete --name "monitor" --ignorenotfound
	-fission httptrigger delete --name "monitor" --ignorenotfound
	-fission httptrigger delete --name "counter" --ignorenotfound
	-fission httptrigger delete --name "splitter" --ignorenotfound
	fission fn delete --name "splitter" --ignorenotfound
	fission fn delete --name "counter" --ignorenotfound
	fission env delete --name "invoker" --ignorenotfound

cleank8s: 
	kubectl delete -f k8s_yaml --ignore-not-found=true

applyk8s: 
	kubectl apply -f k8s_yaml

build:
	cd fctl && go build -o ../target/fctl .

install:
	rm -f $(INSTALL_PATH)
	cp target/fctl $(INSTALL_PATH)

load: 
	fctl load -w ~/the_great_gatsby.txt -b 127.0.0.1:30092 -t word_count_source
