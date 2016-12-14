#!/usr/bin/env bash

go install src/hackathon.2016/sharingservice/datastore/DataStorageService.go

docker build -f docker/DataStorageService_Dockerfile -t gcr.io/$PROJECT_ID/data_storage_service:v1 .

gcloud docker -- push gcr.io/$PROJECT_ID/data_storage_service:v1

gcloud config set compute/zone us-central1-c

gcloud container clusters create data-storage-service-cluster --scopes cloud-platform --num-nodes=2

gcloud container clusters get-credentials data-storage-service-cluster

kubectl run data-storage-service --image=gcr.io/$PROJECT_ID/data_storage_service:v1