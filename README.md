# Demo for 2018 Kubecon Europe

[![Build Status](https://travis-ci.org/Huang-Wei/kubecon-demo.svg?branch=master)](https://travis-ci.org/Huang-Wei/kubecon-demo)

## Running

Please run this demo on Kubernetes 1.9+ versions.

- Local testing (against minikube) in 3 terminals:

```
./main --kubeconfig=/Users/wei.huang1/.kube/config --v=4 --logtostderr=true --hostname=10.10.10.1
./main --kubeconfig=/Users/wei.huang1/.kube/config --v=4 --logtostderr=true --hostname=10.10.10.2
./main --kubeconfig=/Users/wei.huang1/.kube/config --v=4 --logtostderr=true --hostname=10.10.10.3
```
