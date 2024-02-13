#!/bin/bash

for n in $(kubectl get namespaces -o jsonpath={..metadata.name}); do
  kubectl delete --ignore-not-found --all --namespace=$n prometheus,servicemonitor,podmonitor,alertmanager
done

curl -sL https://github.com/prometheus-operator/prometheus-operator/releases/download/v0.71.2/bundle.yaml | kubectl delete --ignore-not-found -f -

for n in $(kubectl get namespaces -o jsonpath={..metadata.name}); do
  kubectl delete --ignore-not-found --namespace=$n service prometheus-operated alertmanager-operated
done

kubectl delete --ignore-not-found customresourcedefinitions \
  prometheuses.monitoring.coreos.com \
  servicemonitors.monitoring.coreos.com \
  podmonitors.monitoring.coreos.com \
  alertmanagers.monitoring.coreos.com \
  prometheusrules.monitoring.coreos.com
