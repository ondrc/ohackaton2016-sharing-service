#!/usr/bin/env bash

kubectl delete service,deployment booking-service

gcloud container clusters delete booking-service-cluster

gsutil rm -r gs://artifacts.$PROJECT_ID.appspot.com/