INSTALL_PATH=/opt/svt/bin/fctl

all: cleanall applyk8s installmemcached installprom installgrafana build install

cleanall: clean cleanfissionmonitor cleanfissionjoin cleanfissionenv cleank8s cleanprom cleangrafana cleanredis cleanmemcached

clean:
	mkdir -p target 
	rm -rf target/*
	rm -f $(INSTALL_PATH)

cleanfissionmonitor:
	fission fn delete --name "monitor" --ignorenotfound
	-fission httptrigger delete --name "monitor" --ignorenotfound

cleanfissonwordcount:
	-fission httptrigger delete --name "counter" --ignorenotfound
	-fission httptrigger delete --name "splitter" --ignorenotfound
	fission fn delete --name "splitter" --ignorenotfound
	fission fn delete --name "counter" --ignorenotfound

cleanfissionjoin:
	-fission httptrigger delete --name "orderlineparse" --ignorenotfound
	-fission httptrigger delete --name "orderlinejoin" --ignorenotfound
	fission fn delete --name "orderlineparse" --ignorenotfound
	fission fn delete --name "orderlinejoin" --ignorenotfound

cleanfissionenv:
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

loadwordcount: 
	fctl load -u "https://www.gutenberg.org/cache/epub/4280/pg4280.txt" -n "the_critique_of_pure_reason.txt" -t word_count_source

loadjoin: 
	fctl load -u "https://media.githubusercontent.com/media/TTraveller7/invokerlib-examples/main/resources/order-line.csv" -n "order-line.csv" -t orderline_source

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

# grafana

cleangrafana:
	kubectl delete -f k8s_yaml/grafana/grafana.yaml --ignore-not-found

installgrafana:
	kubectl apply -f k8s_yaml/grafana/grafana.yaml

# redis

cleanredis:
	kubectl delete -f k8s_yaml/state-redis-stack.yaml --ignore-not-found

installredis:
	kubectl apply -f k8s_yaml/state-redis-stack.yaml

# memcached

cleanmemcached:
	kubectl delete -f k8s_yaml/state-memcached-stack.yaml --ignore-not-found

installmemcached:
	kubectl apply -f k8s_yaml/state-memcached-stack.yaml
