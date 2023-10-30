/*
 * @FilePath: /Users/zengshenglong/Code/GoWorkSpace/operators/es-operator/pkg/k8s/sts.go
 * @Author: kbsonlong kbsonlong@gmail.com
 * @Date: 2023-10-10 11:21:56
 * @LastEditors: kbsonlong kbsonlong@gmail.com
 * @LastEditTime: 2023-10-30 16:20:47
 * @Description:
 * Copyright (c) 2023 by kbsonlong, All Rights Reserved.
 */
package k8s

import (
	"context"
	"fmt"
	"math"
	"os"

	dbv1 "github.com/kbsonlong/es-operator/api/v1"
	apps "k8s.io/api/apps/v1"
	k8scorev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	NameLabel              = "app.kubernetes.io/name"
	InstanceLabel          = "app.kubernetes.io/instance"
	ManagedByLabel         = "app.kubernetes.io/managed-by"
	PartOfLabel            = "app.kubernetes.io/part-of"
	ComponentLabel         = "app.kubernetes.io/component"
	minJvmGB       float64 = 0.5
	gigaByte       int64   = 1 << 30
	megabytes      int64   = 1 << 29
)

func ReconcileStatefulSet(ctx context.Context, es *dbv1.Elasticsearch, req ctrl.Request, c client.Client, Scheme *runtime.Scheme) error {

	log := log.FromContext(ctx)
	JVM_SIZE := getJvmSizeGB(es.Spec.Resources.Limits, true)

	ClusterDomain := dbv1.DefaultDomain
	if ClusterDomain := os.Getenv("ClusterDomain"); ClusterDomain == "" {
		ClusterDomain = dbv1.DefaultDomain
	}

	storageclass := "standard"

	fileMode := int32(0644)
	volumes := []k8scorev1.Volume{
		EnsureVolume("elasticsearch-config", k8scorev1.VolumeSource{
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
		EnsureVolume("plugins", k8scorev1.VolumeSource{
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
				{
					Name:  "ClusterDomain",
					Value: ClusterDomain,
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
					Name:      es.Name,
					MountPath: "/usr/share/elasticsearch/data",
				},
				{
					Name:      "elasticsearch-config",
					MountPath: "/usr/share/elasticsearch/config/elasticsearch.yml",
					SubPath:   "elasticsearch.yml",
				},
				{
					Name:      "plugins",
					MountPath: "/usr/share/elasticsearch/plugins",
					SubPath:   "plugins",
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
			Name:            "sysctl-init",
			Image:           "busybox",
			ImagePullPolicy: "IfNotPresent",
			Command:         []string{"sysctl", "-w", "vm.max_map_count=262144"},
			SecurityContext: &k8scorev1.SecurityContext{
				Privileged: &priObj,
			},
		},
		{
			Name:            "plugins-init",
			Image:           "registry.cn-hangzhou.aliyuncs.com/seam/es-plugins:7.3.2",
			ImagePullPolicy: "IfNotPresent",
			Command:         []string{"/bin/cp", "-r", "/usr/share/elasticsearch/plugins", "/usr/elasticsearch/plugins"},
			VolumeMounts: []k8scorev1.VolumeMount{
				{
					Name:      "plugins",
					MountPath: "/usr/elasticsearch/plugins",
				},
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
				MatchLabels: Labels(es),
			},

			Template: k8scorev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: Labels(es),
				},
				Spec: k8scorev1.PodSpec{
					Affinity:       &es.Spec.Affinity,
					InitContainers: initContainers,
					Containers:     containers,
					Volumes:        volumes,
				},
			},
			VolumeClaimTemplates: []k8scorev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: es.Namespace,
						Name:      es.Name,
					},
					Spec: k8scorev1.PersistentVolumeClaimSpec{
						AccessModes: []k8scorev1.PersistentVolumeAccessMode{
							k8scorev1.ReadWriteOnce,
						},
						Resources: k8scorev1.ResourceRequirements{
							Requests: k8scorev1.ResourceList{
								k8scorev1.ResourceStorage: resource.MustParse("1Gi"),
							},
						},
						StorageClassName: &storageclass,
					},
				},
			},
		},
	}

	// statefulset 与 crd 资源建立关联,
	// 建立关联后，删除 crd 资源时就会将 statefulset 也删除掉
	log.Info("set sts reference")
	if err := controllerutil.SetControllerReference(es, statefulset, Scheme); err != nil {
		log.Error(err, "SetControllerReference error")
		return err
	}

	if err := c.Create(ctx, statefulset); err != nil {
		if errors.IsAlreadyExists(err) {
			if err := c.Update(ctx, statefulset); err != nil {
				log.Error(err, "create service error")
				return err
			}
		}
	}

	return nil
}

func EnsureVolume(name string, source k8scorev1.VolumeSource) k8scorev1.Volume {
	return k8scorev1.Volume{
		Name:         name,
		VolumeSource: source,
	}
}

func Labels(es *dbv1.Elasticsearch) map[string]string {
	return map[string]string{
		NameLabel:      "es-cluster",
		InstanceLabel:  es.Name,
		ManagedByLabel: "es-operator",
		PartOfLabel:    "es-cluster",
	}
}

// 0.5 * ( memoryLimit -  1GB )
func getJvmSizeGB(resourceList k8scorev1.ResourceList, subtract1GB bool) string {
	maxMemory := resourceList[k8scorev1.ResourceMemory]

	var size float64
	if subtract1GB {
		size = math.Floor(0.5 * float64(maxMemory.Value()-gigaByte))
	} else {
		size = math.Floor(0.5 * float64(maxMemory.Value()))
	}
	sizeGB := size / float64(gigaByte)
	if sizeGB < minJvmGB {
		sizeGB = minJvmGB
	}

	jvm := fmt.Sprintf("%dG", int(size))
	// fmt.Println(size / float64(megabytes))
	if sizeGB < 1 {
		sizeMB := sizeGB * 1024
		jvm = fmt.Sprintf("%dM", int(sizeMB))
	}

	return jvm
}

func GetPods(ctx context.Context, cluster *dbv1.Elasticsearch, c client.Client) (k8scorev1.PodList, error) {
	Pods := k8scorev1.PodList{}
	err := c.List(ctx,
		&Pods,
		&client.ListOptions{
			Namespace:     cluster.Namespace,
			LabelSelector: labels.SelectorFromSet(Labels(cluster)),
		},
	)
	return Pods, err
}
