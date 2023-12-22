/*
 * @FilePath: /Users/zengshenglong/Code/GoWorkSpace/operators/es-operator/pkg/k8s/configmap.go
 * @Author: kbsonlong kbsonlong@gmail.com
 * @Date: 2023-10-10 11:22:36
 * @LastEditors: kbsonlong kbsonlong@gmail.com
 * @LastEditTime: 2023-12-22 16:50:55
 * @Description:
 * Copyright (c) 2023 by kbsonlong, All Rights Reserved.
 */
package k8s

import (
	"context"
	"encoding/json"
	"fmt"

	jsonpatch "github.com/evanphx/json-patch"
	dbv1 "github.com/kbsonlong/es-operator/api/v1"
	"gopkg.in/yaml.v2"
	k8scorev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func ReconcileConfigMap(ctx context.Context, es *dbv1.Elasticsearch, req ctrl.Request, c client.Client, Scheme *runtime.Scheme) error {
	log := log.FromContext(ctx)

	// 	tempMap := make(map[interface{}]interface{})
	// 	tempMap["Name"] = es.Name
	// 	tempMap["Size"] = int32(es.Spec.Size)
	// 	temp := `cluster.name: {{.Name}}
	// cluster.initial_master_nodes: [{{range $x := loop .Name .Size}}{{$x}},{{- end}}]
	// discovery.seed_hosts: ["${HEADLESS_SERVICE_NAME}.${NAMESPACE}.svc.${ClusterDomain}"]
	// network.host: "0"
	// network.bind_host: ${POD_IP}
	// network.publish_host: ${POD_IP}
	// node.name: ${POD_NAME}
	// node.attr.k8s_node_name: ${NODE_NAME}
	// http.port:  9200
	// transport.port:  9300
	// path.data: /usr/share/elasticsearch/data
	// path.logs: /usr/share/elasticsearch/logs
	// bootstrap.memory_lock: false
	// bootstrap.system_call_filter: false
	// cluster.routing.allocation.same_shard.host: true
	// indices.query.bool.max_clause_count : 2048
	// indices.memory.index_buffer_size: 30%
	// indices.fielddata.cache.size: 40%
	// indices.breaker.fielddata.limit: 70%
	// indices.recovery.max_bytes_per_sec: 20mb
	// indices.breaker.total.use_real_memory: false
	// thread_pool.write.queue_size: 1000
	// xpack.security.enabled: false
	// xpack.security.transport.ssl.enabled: false
	// xpack.security.http.ssl.enabled: false
	// `
	// data := ParseConf(temp, tempMap)
	data, _ := MergePatchYAML(es)

	cm := &k8scorev1.ConfigMap{}

	cm.TypeMeta = metav1.TypeMeta{
		Kind:       "ConfigMap",
		APIVersion: "v1",
	}
	cm.ObjectMeta = metav1.ObjectMeta{
		Name:      fmt.Sprintf("%s-config", es.Name),
		Namespace: es.Namespace,
	}
	cm.Data = map[string]string{
		// "elasticsearch.yml": data.String(),
		"elasticsearch.yml": string(data),
	}
	// 建立关联后，删除 crd 资源时就会将 service 也删除掉
	// log.Info("set svc reference")
	if err := controllerutil.SetControllerReference(es, cm, Scheme); err != nil {
		log.Error(err, "SetControllerReference error")
		return err
	}
	if err := c.Create(ctx, cm); err != nil {
		if errors.IsAlreadyExists(err) {
			if err := c.Update(ctx, cm); err != nil {
				log.Error(err, "Create ConfigMap error")
				return err
			}
		}
	}
	return nil
}

var nodeAttrK8sNodeName = "k8s_node_name"
var nodeAttrNodeName = fmt.Sprintf("%s.%s", dbv1.NodeAttr, nodeAttrK8sNodeName)

// MarshalYAML implements the Marshaler interface.
func MergePatchYAML(es *dbv1.Elasticsearch) ([]byte, error) {
	cfg := map[string]interface{}{
		// derive node name dynamically from the pod name, injected as env var
		dbv1.NodeName:    "${POD_NAME}",
		dbv1.ClusterName: es.Name,
		// use the DNS name as the publish host
		dbv1.NetworkPublishHost: "${POD_IP}",
		dbv1.HTTPPublishHost:    "${POD_NAME}.${HEADLESS_SERVICE_NAME}.${NAMESPACE}.svc",
		dbv1.NetworkHost:        "0",
		// allow ES to be aware of k8s node the pod is running on when allocating shards
		dbv1.ShardAwarenessAttributes: dbv1.NodeAttr,
		nodeAttrNodeName:              "${NODE_NAME}",
	}
	dstJSON, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}
	// 序列化补丁结构体到JSON，这个补丁描述了如何修改目标（原始）对象
	patchJSON, err := json.Marshal(es.Spec.Config)
	if err != nil {
		return nil, err
	}

	// 使用补丁合并目标（原始）对象
	mergedJSON, err := jsonpatch.MergePatch(dstJSON, patchJSON)
	if err != nil {
		return nil, err
	}
	data := &map[string]interface{}{}
	if err := json.Unmarshal(mergedJSON, data); err != nil {
		fmt.Println(err)
	}

	return yaml.Marshal(data)

}
