#!/bin/bash
export FISSION_NAMESPACE="fission"
kubectl create namespace $FISSION_NAMESPACE
kubectl create -k "github.com/fission/fission/crds/v1?ref=v1.20.0"
helm install --version v1.20.0 --namespace $FISSION_NAMESPACE fission \
  --set serviceType=NodePort,routerServiceType=NodePort \
  fission-charts/fission-all
export FISSION_ROUTER="http://127.0.0.1:$(kubectl -n fission get svc router -o jsonpath='{...nodePort}')"
