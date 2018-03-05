#!/usr/bin/env bash

for i in 1 2 3; do
    kubectl create -f foo$i.yaml
done

sleep 3

for i in 1 2 3; do
    kubectl delete -f foo$i.yaml
done
