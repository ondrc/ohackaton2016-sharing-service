#!/usr/bin/env bash

kubectl delete service,deployment data-storage-service

gcloud -q container clusters delete data-storage-service-cluster

gsutil rm -r gs://artifacts.$PROJECT_ID.appspot.com/