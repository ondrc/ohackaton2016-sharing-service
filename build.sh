#!/usr/bin/env bash

echo ''
echo '*******************************************'
echo '*      Setting environment variables      *'
echo '*******************************************'
echo ''
. setenv.sh

echo ''
echo '*******************************************'
echo '*        Fetching code dependencies       *'
echo '*******************************************'
echo ''
go get hackathon.2016/sharingservice/common
go get hackathon.2016/sharingservice/datastore
go get hackathon.2016/sharingservice/query
go get hackathon.2016/sharingservice/registration

echo ''
echo '*******************************************'
echo '*      Deploying data storage service     *'
echo '*******************************************'
echo ''
bash build/deploy_data_storage_service.sh

echo ''
echo '*******************************************'
echo '*         Deploying query service         *'
echo '*******************************************'
echo ''
bash build/deploy_query_service.sh

echo ''
echo '*******************************************'
echo '*      Deploying registration service     *'
echo '*******************************************'
echo ''
bash build/deploy_registration_service.sh

echo ''
echo '*******************************************'
echo '*         Waiting for service IPs         *'
echo '*******************************************'
echo ''
sleep 1m
bash build/get_service_ips.sh

echo ''
echo '*******************************************'
echo '*          Successfully deployed          *'
echo '*******************************************'
echo ''