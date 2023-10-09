/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"bytes"
	"context"
	"fmt"
	"text/template"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	dbv1 "github.com/kbsonlong/es-operator/api/v1"
	apps "k8s.io/api/apps/v1"
	k8scorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ElasticsearchReconciler reconciles a Elasticsearch object
type ElasticsearchReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

const (
	NameLabel      = "app.kubernetes.io/name"
	InstanceLabel  = "app.kubernetes.io/instance"
	ManagedByLabel = "app.kubernetes.io/managed-by"
	PartOfLabel    = "app.kubernetes.io/part-of"
	ComponentLabel = "app.kubernetes.io/component"
)

//+kubebuilder:rbac:groups=db.alongparty.cn,resources=elasticsearches,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=db.alongparty.cn,resources=elasticsearches/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=db.alongparty.cn,resources=elasticsearches/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=replicasets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Elasticsearch object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *ElasticsearchReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// TODO(user): your logic here
	es := &dbv1.Elasticsearch{}
	err := r.Get(ctx, req.NamespacedName, es)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Return and don't requeue
			logger.Info("ES Cluster resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get ES Cluster")
		return ctrl.Result{RequeueAfter: time.Second * 5}, err
	}

	r.doReconcileConfigMap(ctx, es, req)
	r.doReconcileServices(ctx, es)
	r.doReconcileStatefulSet(ctx, es)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ElasticsearchReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dbv1.Elasticsearch{}).
		Complete(r)
}

func (r *ElasticsearchReconciler) doReconcileStatefulSet(ctx context.Context, es *dbv1.Elasticsearch) error {

	log := log.FromContext(ctx)
	JVM_SIZE := "800m"
	// MemLimit = es.Spec.Resources.Limits.Memory()

	fileMode := int32(0644)
	volumes := []k8scorev1.Volume{
		ensureVolume("elasticsearch-config", k8scorev1.VolumeSource{
			ConfigMap: &k8scorev1.ConfigMapVolumeSource{
				LocalObjectReference: k8scorev1.LocalObjectReference{
					Name: fmt.Sprintf("%s-config", es.Name),
				},
				Items: []k8scorev1.KeyToPath{
					{
						Key:  "elasticsearch.yml",
						Path: "elasticsearch.yml",
					},
				},
				DefaultMode: &fileMode,
			},
		}),
		ensureVolume("data", k8scorev1.VolumeSource{
			EmptyDir: &k8scorev1.EmptyDirVolumeSource{},
		}),
	}
	priObj := bool(true)
	runUser := int64(1000)
	containers := []k8scorev1.Container{
		{
			Name:            es.Name,
			Image:           es.Spec.Image,
			ImagePullPolicy: "IfNotPresent",
			Resources:       es.Spec.Resources,
			SecurityContext: &k8scorev1.SecurityContext{
				Privileged: &priObj,
				RunAsUser:  &runUser,
				Capabilities: &k8scorev1.Capabilities{
					Add: []k8scorev1.Capability{
						"IPC_LOCK",
						"SYS_RESOURCE",
					},
				},
			},
			Env: []k8scorev1.EnvVar{
				{
					Name: "POD_IP",
					ValueFrom: &k8scorev1.EnvVarSource{
						FieldRef: &k8scorev1.ObjectFieldSelector{
							FieldPath: "status.podIP",
						},
					},
				},
				{
					Name: "POD_NAME",
					ValueFrom: &k8scorev1.EnvVarSource{
						FieldRef: &k8scorev1.ObjectFieldSelector{
							FieldPath: "metadata.name",
						},
					},
				},
				{
					Name: "NODE_NAME",
					ValueFrom: &k8scorev1.EnvVarSource{
						FieldRef: &k8scorev1.ObjectFieldSelector{
							FieldPath: "spec.nodeName",
						},
					},
				},
				{
					Name: "NAMESPACE",
					ValueFrom: &k8scorev1.EnvVarSource{
						FieldRef: &k8scorev1.ObjectFieldSelector{
							FieldPath: "metadata.namespace",
						},
					},
				},
				{
					Name:  "ES_JAVA_OPTS",
					Value: fmt.Sprintf("-Xms%s -Xmx%s", JVM_SIZE, JVM_SIZE),
				},
				{
					Name:  "HEADLESS_SERVICE_NAME",
					Value: fmt.Sprintf("%s-headless", es.Name),
				},
			},
			Ports: []k8scorev1.ContainerPort{
				{
					Name:          "http",
					ContainerPort: 9200,
				},
				{
					Name:          "transport",
					ContainerPort: 9300,
				},
			},
			VolumeMounts: []k8scorev1.VolumeMount{
				{
					Name:      "data",
					MountPath: "/usr/share/elasticsearch/data",
				},
				{
					Name:      "elasticsearch-config",
					MountPath: "/usr/share/elasticsearch/config/elasticsearch.yml",
					SubPath:   "elasticsearch.yml",
				},
			},
			ReadinessProbe: &k8scorev1.Probe{
				ProbeHandler: k8scorev1.ProbeHandler{
					HTTPGet: &k8scorev1.HTTPGetAction{
						Path: "/_cluster/health?local=true",
						Port: intstr.IntOrString{
							IntVal: 9200,
						},
						Scheme: "HTTP",
					},
				},
			},
			// Command: []string{"/bin/sh", "-c", "sleep 3600"},
		},
	}

	initContainers := []k8scorev1.Container{
		{
			Name:            fmt.Sprintf("%s-init", es.Name),
			Image:           "busybox",
			ImagePullPolicy: "IfNotPresent",
			Command:         []string{"sysctl", "-w", "vm.max_map_count=262144"},
			SecurityContext: &k8scorev1.SecurityContext{
				Privileged: &priObj,
			},
		},
	}

	statefulset := &apps.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: es.Namespace,
			Name:      es.Name,
		},
		Spec: apps.StatefulSetSpec{
			Replicas: &es.Spec.Size,
			Selector: &metav1.LabelSelector{
				MatchLabels: r.Labels(es),
			},

			Template: k8scorev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: r.Labels(es),
				},
				Spec: k8scorev1.PodSpec{
					InitContainers: initContainers,
					Containers:     containers,
					Volumes:        volumes,
				},
			},
		},
	}

	// statefulset 与 crd 资源建立关联,
	// 建立关联后，删除 crd 资源时就会将 statefulset 也删除掉
	log.Info("set sts reference")
	if err := controllerutil.SetControllerReference(es, statefulset, r.Scheme); err != nil {
		log.Error(err, "SetControllerReference error")
		return err
	}
	err := r.Create(ctx, statefulset)
	if err != nil {
		return err
	}
	return nil
}

func (r *ElasticsearchReconciler) Labels(es *dbv1.Elasticsearch) map[string]string {
	return map[string]string{
		NameLabel:      "es-cluster",
		InstanceLabel:  es.Name,
		ManagedByLabel: "es-operator",
		PartOfLabel:    "es-cluster",
	}
}

func (r *ElasticsearchReconciler) doReconcileServices(ctx context.Context, es *dbv1.Elasticsearch) error {
	log := log.FromContext(ctx)
	// log := r.Log.WithValues("func", "doReconcileService")
	selector := r.Labels(es)
	svc := &k8scorev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: es.Namespace,
			Name:      es.Name,
			Labels:    selector,
		},
		Spec: k8scorev1.ServiceSpec{
			Ports: []k8scorev1.ServicePort{
				{
					Name:     "http",
					Port:     9200,
					Protocol: k8scorev1.ProtocolTCP,
				},
			},
			Selector: selector,
			Type:     k8scorev1.ServiceTypeClusterIP,
		},
	}

	// service 与 crd 资源建立关联,
	// 建立关联后，删除 crd 资源时就会将 service 也删除掉
	// log.Info("set svc reference")
	if err := controllerutil.SetControllerReference(es, svc, r.Scheme); err != nil {
		log.Error(err, "SetControllerReference error")
		return err
	}
	// 创建service
	// log.Info("start create service")
	if err := r.Create(ctx, svc); err != nil {
		if errors.IsAlreadyExists(err) {
			if err := r.Update(ctx, svc); err != nil {
				log.Error(err, "create service error")
				return err
			}
		}

		// return err
	}

	// 创建 headless service
	headlessService := &k8scorev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: es.Namespace,
			Name:      fmt.Sprintf("%s-headless", es.Name),
			Labels:    selector,
		},
		Spec: k8scorev1.ServiceSpec{
			Ports: []k8scorev1.ServicePort{
				{
					Name:     "transport",
					Port:     9300,
					Protocol: k8scorev1.ProtocolTCP,
				},
			},
			Selector:  selector,
			Type:      k8scorev1.ServiceTypeClusterIP,
			ClusterIP: "None",
		},
	}

	// service 与 crd 资源建立关联,
	// 建立关联后，删除 crd 资源时就会将 service 也删除掉
	// log.Info("set reference")
	if err := controllerutil.SetControllerReference(es, headlessService, r.Scheme); err != nil {
		log.Error(err, "SetControllerReference error")
		return err
	}

	// log.Info("start create headlessService")
	if err := r.Create(ctx, headlessService); err != nil {
		if errors.IsAlreadyExists(err) {
			if err := r.Update(ctx, headlessService); err != nil {
				log.Error(err, "create headlessService error")
				return err
			}
		}
	}
	return nil
}

func (r *ElasticsearchReconciler) doReconcileConfigMap(ctx context.Context, es *dbv1.Elasticsearch, req ctrl.Request) error {
	log := log.FromContext(ctx)
	cm := &k8scorev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-config", es.Name),
			Namespace: es.Namespace,
		},
	}

	tempMap := make(map[interface{}]interface{})
	tempMap["Name"] = es.Name
	tempMap["Size"] = int32(es.Spec.Size)

	temp := `cluster:
  name: {{.Name}}
  routing:
    allocation:
      awareness:
        attributes: k8s_node_name
  initial_master_nodes: [{{range $x := loop .Name .Size}}{{$x}},{{- end}}]
bootstrap:
  memory_lock: false
discovery:
  seed_hosts: ["${HEADLESS_SERVICE_NAME}.${NAMESPACE}.svc.cluster.local"]
http:
  publish_host: ${POD_IP}
network:
  host: "0"
  publish_host: ${POD_IP}
node:
  attr:
    k8s_node_name: ${NODE_NAME}
  name: ${POD_NAME}
  store:
    allow_mmap: false
path:
  data: /usr/share/elasticsearch/data
  logs: /usr/share/elasticsearch/logs
indices:
  query:
    bool:
      max_nested_depth: 500
      max_clause_count: 2048
  memory:
    index_buffer_size: 30%
  fielddata:
    cache:
      size: 40%
  breaker:
    fielddata:
      limit: 70%
    total:
      use_real_memory: false
  recovery:
    max_bytes_per_sec: 20mb
thread_pool:
  write:
    queue_size: 1000
xpack:
  security:
    enabled: false
    transport:
      ssl:
        enabled: false
    http:
      ssl:
        enabled: false
`

	data := parseConf(temp, tempMap)
	err := r.Client.Get(ctx, req.NamespacedName, cm)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Create ConfigMap")
			cm.Data = map[string]string{
				"elasticsearch.yml": data.String(),
			}

			err = r.Client.Create(ctx, cm)
			return err
		}
		fmt.Println(err)
		return err
	}

	cm.Data = map[string]string{
		"elasticsearch.yml": data.String(),
	}
	err = r.Client.Update(ctx, cm)
	if err != nil {
		log.Info("Update ConfigMap failed")
		return err
	}
	return nil
}

func parseConf(conftemp string, es map[interface{}]interface{}) bytes.Buffer {
	funcMap := template.FuncMap{
		"loop": func(name string, to int32) <-chan string {
			ch := make(chan string)
			go func() {
				for i := 0; i <= int(to)-1; i++ {
					ch <- fmt.Sprintf("%s-%d", name, i)
				}
				close(ch)
			}()
			return ch
		},
	}

	tmpl, err := template.New("conf").Funcs(funcMap).Parse(conftemp)
	if err != nil {
		fmt.Println(err)
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, es)
	if err != nil {
		panic(err)
	}
	if err != nil {
		fmt.Println(err)
	}
	return buf
}

func ensureVolume(name string, source k8scorev1.VolumeSource) k8scorev1.Volume {
	return k8scorev1.Volume{
		Name:         name,
		VolumeSource: source,
	}
}
