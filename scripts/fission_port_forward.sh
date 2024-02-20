#!/bin/bash
kubectl --namespace fission port-forward $(kubectl --namespace fission get pod -l svc=router -o name) 31314:8888 &
