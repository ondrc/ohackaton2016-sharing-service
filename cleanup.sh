#!/usr/bin/env bash

echo ''
echo '*******************************************'
echo '*      Setting environment variables      *'
echo '*******************************************'
echo ''
. setenv.sh

echo ''
echo '*******************************************'
echo '*        Removing booking service         *'
echo '*******************************************'
echo ''
bash build/cleanup_booking_service.sh

echo ''
echo '*******************************************'
echo '*      Removing registration service      *'
echo '*******************************************'
echo ''
bash build/cleanup_registration_service.sh

echo ''
echo '*******************************************'
echo '*         Removing query service          *'
echo '*******************************************'
echo ''
bash build/cleanup_query_service.sh

echo ''
echo '*******************************************'
echo '*      Removing data storage service      *'
echo '*******************************************'
echo ''
bash build/cleanup_data_storage_service.sh

echo ''
echo '*******************************************'
echo '*         Successfully undeployed         *'
echo '*******************************************'
echo ''
