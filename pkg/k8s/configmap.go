/*
 * @FilePath: /pkg/k8s/configmap.go
 * @Author: kbsonlong kbsonlong@gmail.com
 * @Date: 2023-10-10 11:22:36
 * @LastEditors: kbsonlong kbsonlong@gmail.com
 * @LastEditTime: 2023-10-11 10:43:33
 * @Description:
 * Copyright (c) 2023 by kbsonlong, All Rights Reserved.
 */
package k8s

import (
	"context"
	"fmt"

	dbv1 "github.com/kbsonlong/es-operator/api/v1"
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

	tempMap := make(map[interface{}]interface{})
	tempMap["Name"] = es.Name
	tempMap["Size"] = int32(es.Spec.Size)
	temp := `bootstrap.memory_lock: false
bootstrap.system_call_filter: true
cluster.routing.allocation.same_shard.host: true
cluster.name: {{.Name}}
cluster.initial_master_nodes: [{{range $x := loop .Name .Size}}{{$x}},{{- end}}]
discovery.seed_hosts: ["${HEADLESS_SERVICE_NAME}.${NAMESPACE}.svc.cluster.local"]
discovery.zen.ping_timeout: 90s
discovery.zen.fd.ping_interval: 10s
discovery.zen.fd.ping_timeout: 120s
discovery.zen.fd.ping_retries: 12
network.host: "0"
network.bind_host: ${POD_IP}
network.publish_host: ${POD_IP}
node.master: true
node.data: true
node.name: ${POD_NAME}
node.attr.k8s_node_name: ${NODE_NAME}
http.port:  9200
transport.tcp.port:  9300
path.data: /usr/share/elasticsearch/data
path.logs: /usr/share/elasticsearch/logs
indices.query.bool.max_clause_count : 2048
indices.memory.index_buffer_size: 30%
indices.fielddata.cache.size: 40%
indices.breaker.fielddata.limit: 70%
indices.recovery.max_bytes_per_sec: 20mb
indices.breaker.total.use_real_memory: false
thread_pool.write.queue_size: 1000
`
	data := ParseConf(temp, tempMap)
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
		"elasticsearch.yml": data.String(),
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
