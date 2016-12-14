#!/usr/bin/env bash

go install src/hackathon.2016/sharingservice/registration/RegistrationService.go

docker build -f docker/RegistrationService_Dockerfile -t gcr.io/$PROJECT_ID/registration_service:v1 .

gcloud docker -- push gcr.io/$PROJECT_ID/registration_service:v1

gcloud config set compute/zone us-central1-c

gcloud container clusters create registration-service-cluster --scopes cloud-platform --num-nodes=2

gcloud container clusters get-credentials registration-service-cluster

kubectl run registration-service --image=gcr.io/$PROJECT_ID/registration_service:v1 --port=8081

kubectl expose deployment registration-service --type="LoadBalancer"