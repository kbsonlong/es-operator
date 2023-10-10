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

func ReconcileDeploymant(ctx context.Context, kb *dbv1.Kibana, req ctrl.Request, c client.Client, Scheme *runtime.Scheme) error {

	log := log.FromContext(ctx)

	selector := map[string]string{
		NameLabel:      "es-cluster",
		InstanceLabel:  kb.Name,
		ManagedByLabel: "es-operator",
		PartOfLabel:    "es-cluster",
	}
	// priObj := bool(true)
	// runUser := int64(1000)
	containers := []k8scorev1.Container{
		{
			Name:            kb.Name,
			Image:           kb.Spec.Image,
			ImagePullPolicy: "IfNotPresent",
			Resources:       kb.Spec.Resources,
			// SecurityContext: &k8scorev1.SecurityContext{
			// 	Privileged: &priObj,
			// 	RunAsUser:  &runUser,
			// 	Capabilities: &k8scorev1.Capabilities{
			// 		Add: []k8scorev1.Capability{
			// 			"IPC_LOCK",
			// 			"SYS_RESOURCE",
			// 		},
			// 	},
			// },
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
					Name:  "ELASTICSEARCH_HOSTS",
					Value: fmt.Sprintf("[\"http://elasticsearch-sample.%s.svc.cluster.local:9200\"]", kb.Namespace),
				},
			},
			Ports: []k8scorev1.ContainerPort{
				{
					Name:          "http",
					ContainerPort: 5601,
				},
			},
			ReadinessProbe: &k8scorev1.Probe{
				ProbeHandler: k8scorev1.ProbeHandler{
					HTTPGet: &k8scorev1.HTTPGetAction{
						Path: "/app/home#/",
						Port: intstr.IntOrString{
							IntVal: 5601,
						},
						Scheme: "HTTP",
					},
				},
			},
		},
	}

	deployment := &apps.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: kb.Namespace,
			Name:      kb.Name,
		},
		Spec: apps.DeploymentSpec{
			Replicas: &kb.Spec.Size,
			Selector: &metav1.LabelSelector{
				MatchLabels: selector,
			},

			Template: k8scorev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: selector,
				},
				Spec: k8scorev1.PodSpec{
					Containers: containers,
				},
			},
		},
	}

	// statefulset 与 crd 资源建立关联,
	// 建立关联后，删除 crd 资源时就会将 statefulset 也删除掉
	log.Info("set sts reference")
	if err := controllerutil.SetControllerReference(kb, deployment, Scheme); err != nil {
		log.Error(err, "SetControllerReference error")
		return err
	}

	if err := c.Create(ctx, deployment); err != nil {
		if errors.IsAlreadyExists(err) {
			if err := c.Update(ctx, deployment); err != nil {
				log.Error(err, "create service error")
				return err
			}
		}
		return err
	}

	return nil
}
