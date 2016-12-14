#!/usr/bin/env bash

go install src/hackathon.2016/sharingservice/booking/BookingService.go

docker build -f docker/BookingService_Dockerfile -t gcr.io/$PROJECT_ID/booking_service:v1 .

gcloud docker -- push gcr.io/$PROJECT_ID/booking_service:v1

gcloud config set compute/zone us-central1-c

gcloud container clusters create booking-service-cluster --scopes cloud-platform

gcloud container clusters get-credentials booking-service-cluster

kubectl run booking-service --image=gcr.io/$PROJECT_ID/booking_service:v1 --port=8083

kubectl expose deployment booking-service --type="LoadBalancer"

sleep 1m
kubectl get services booking-service