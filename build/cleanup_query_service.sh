#!/usr/bin/env bash

kubectl delete service,deployment query-service

gcloud -q container clusters delete query-service-cluster

gsutil rm -r gs://artifacts.$PROJECT_ID.appspot.com/