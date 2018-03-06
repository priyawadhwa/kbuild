#!/bin/bash
set -e

kubectl create configmap dockerfile-config --from-file=Dockerfile