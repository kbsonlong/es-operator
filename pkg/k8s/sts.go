/*
 * @FilePath: /pkg/k8s/sts.go
 * @Author: kbsonlong kbsonlong@gmail.com
 * @Date: 2023-10-10 11:21:56
 * @LastEditors: kbsonlong kbsonlong@gmail.com
 * @LastEditTime: 2023-10-10 12:02:49
 * @Description:
 * Copyright (c) 2023 by kbsonlong, All Rights Reserved.
 */
package k8s

import (
	"context"
	"fmt"

	dbv1 "github.com/kbsonlong/es-operator/api/v1"
	apps "k8s.io/api/apps/v1"
	k8scorev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	NameLabel      = "app.kubernetes.io/name"
	InstanceLabel  = "app.kubernetes.io/instance"
	ManagedByLabel = "app.kubernetes.io/managed-by"
	PartOfLabel    = "app.kubernetes.io/part-of"
	ComponentLabel = "app.kubernetes.io/component"
)

func ReconcileStatefulSet(ctx context.Context, es *dbv1.Elasticsearch, req ctrl.Request, c client.Client, Scheme *runtime.Scheme) error {

	log := log.FromContext(ctx)
	JVM_SIZE := "800m"
	// MemLimit = es.Spec.Resources.Limits.Memory()

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
		EnsureVolume("data", k8scorev1.VolumeSource{
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
				MatchLabels: Labels(es),
			},

			Template: k8scorev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: Labels(es),
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
		return err
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
