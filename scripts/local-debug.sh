#!/bin/bash
###
 # @FilePath: /scripts/local-debug.sh
 # @Author: kbsonlong kbsonlong@gmail.com
 # @Date: 2023-10-11 07:41:24
 # @LastEditors: kbsonlong kbsonlong@gmail.com
 # @LastEditTime: 2023-10-11 08:07:37
 # @Description: 
 # Copyright (c) 2023 by kbsonlong, All Rights Reserved. 
### 

which telepresence
if [[ $? != 0 ]]; then
    echo "Please Install Telepresence on Your Machine"
    echo "https://www.telepresence.io/docs/latest/quick-start/"
fi

kubectl get deploy -n ambassador traffic-manager

if [[ $? == 0 ]]; then
  echo "Telepresence already exist in kind cluster"
else
  echo "Install Telepresence into kind cluster"
  telepresence helm install
fi

i=0
while [ $i -ne 12 ];
do
  kubectl get pod -n ambassador -l app=traffic-manager|grep 'Running'
  if [[ $? == 0 ]]; then
    telepresence connect
    echo "Telepresence connected successfully"
    break
  else
    echo "Waiting install telepresence into kind cluster"
  fi
  sleep 10
done
