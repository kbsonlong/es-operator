#!/bin/bash
###
 # @FilePath: /scripts/create_kind.sh
 # @Author: kbsonlong kbsonlong@gmail.com
 # @Date: 2023-10-11 07:32:56
 # @LastEditors: kbsonlong kbsonlong@gmail.com
 # @LastEditTime: 2023-10-11 08:03:34
 # @Description: 
 # Copyright (c) 2023 by kbsonlong, All Rights Reserved. 
### 

ROOT_DIR=$(dirname $0)/
TEMP_DIR=${ROOT_DIR}/../temp

mkdir -p ${TEMP_DIR}
cd ${TEMP_DIR}

cat << EOF > kind-cluster.yaml
kind: Cluster
apiVersion: "kind.x-k8s.io/v1alpha4"
networking:
  podSubnet: "10.10.0.0/16"
  serviceSubnet: "10.11.0.0/16"
nodes:
  - role: control-plane
    image: registry.cn-hangzhou.aliyuncs.com/seam/node:v1.24.15
EOF

kind get clusters |grep cluster
if [[ $? != 0 ]]; then
  kind create cluster --name cluster --config=kind-cluster.yaml
else
  echo "kind cluster already exist"
fi


