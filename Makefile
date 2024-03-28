INSTALL_PATH=/opt/svt/bin/fctl

all: cleanall applyk8s installprom build install

cleanall: clean cleanfission cleank8s cleanprom

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
	kubectl delete -f k8s_yaml/storage --ignore-not-found=true
	kubectl delete svc -l environmentName=invoker
	kubectl delete deployment -l environmentName=invoker

applyk8s: 
	kubectl apply -f k8s_yaml/storage

build:
	cd pkg/fctl && go build -o ../../target/fctl .

install:
	rm -f $(INSTALL_PATH)
	cp target/fctl $(INSTALL_PATH)
	scripts/fission_port_forward.sh

load: 
	fctl load -u "https://www.gutenberg.org/cache/epub/4280/pg4280.txt" -n "the_critique_of_pure_reason.txt" -t word_count_source

# prometheus

cleanprom:
	kubectl delete -f k8s_yaml/prometheus/operator.yaml --ignore-not-found
	kubectl delete -f k8s_yaml/prometheus/role.yaml --ignore-not-found
	./scripts/clean_prometheus.sh

installprom: 
	curl -sL https://github.com/prometheus-operator/prometheus-operator/releases/download/v0.71.2/bundle.yaml | kubectl create -f -
	kubectl wait --for=condition=Ready pods -l  app.kubernetes.io/name=prometheus-operator -n default
	kubectl apply -f k8s_yaml/prometheus/role.yaml
	kubectl apply -f k8s_yaml/prometheus/operator.yaml
