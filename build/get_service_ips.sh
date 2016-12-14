#!/usr/bin/env bash

echo ''
echo '*******************************************'
echo '*              Query Service              *'
echo '*******************************************'
echo ''
gcloud container clusters get-credentials query-service-cluster
kubectl get services query-service

echo ''
echo '*******************************************'
echo '*          Registration Service           *'
echo '*******************************************'
echo ''
gcloud container clusters get-credentials registration-service-cluster
kubectl get services registration-service