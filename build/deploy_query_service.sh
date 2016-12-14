#!/usr/bin/env bash

go install hackathon.2016/sharingservice/query

docker build -f docker/QueryService_Dockerfile -t gcr.io/$PROJECT_ID/query_service:v1 .

gcloud docker -- push gcr.io/$PROJECT_ID/query_service:v1

gcloud config set compute/zone us-central1-c

gcloud container clusters create query-service-cluster --scopes cloud-platform

gcloud container clusters get-credentials query-service-cluster

kubectl run query-service --image=gcr.io/$PROJECT_ID/query_service:v1 --port=8080

kubectl expose deployment query-service --type="LoadBalancer"

sleep 1m
kubectl get services query-service