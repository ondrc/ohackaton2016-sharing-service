#!/usr/bin/env bash

kubectl delete service,deployment registration-service

gcloud -q container clusters delete registration-service-cluster

gsutil rm -r gs://artifacts.$PROJECT_ID.appspot.com/