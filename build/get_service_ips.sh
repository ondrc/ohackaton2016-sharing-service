#!/usr/bin/env bash

echo ''
echo '*******************************************'
echo '*             Booking Service             *'
echo '*******************************************'
echo ''
kubectl get services booking-service

echo ''
echo '*******************************************'
echo '*              Query Service              *'
echo '*******************************************'
echo ''
kubectl get services query-service

echo ''
echo '*******************************************'
echo '*          Registration Service           *'
echo '*******************************************'
echo ''
kubectl get services registration-service