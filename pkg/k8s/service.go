/*
 * @FilePath: /pkg/k8s/service.go
 * @Author: kbsonlong kbsonlong@gmail.com
 * @Date: 2023-10-10 11:22:17
 * @LastEditors: kbsonlong kbsonlong@gmail.com
 * @LastEditTime: 2023-10-11 10:45:12
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

func ReconcileElasticServices(ctx context.Context, es *dbv1.Elasticsearch, req ctrl.Request, c client.Client, Scheme *runtime.Scheme) error {
	log := log.FromContext(ctx)
	// log := r.Log.WithValues("func", "doReconcileService")
	selector := Labels(es)
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
	if err := controllerutil.SetControllerReference(es, svc, Scheme); err != nil {
		log.Error(err, "SetControllerReference error")
		return err
	}
	// 创建service
	// log.Info("start create service")

	if err := c.Create(ctx, svc); err != nil {
		if errors.IsAlreadyExists(err) {
			if err := c.Update(ctx, svc); err != nil {
				log.Error(err, "create service error")
				return err
			}
		}
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
	if err := controllerutil.SetControllerReference(es, headlessService, Scheme); err != nil {
		log.Error(err, "SetControllerReference error")
		return err
	}

	// log.Info("start create headlessService")
	if err := c.Create(ctx, headlessService); err != nil {
		if errors.IsAlreadyExists(err) {
			if err := c.Update(ctx, headlessService); err != nil {
				log.Error(err, "create headlessService error")
				return err
			}
		}
	}
	return nil
}

func ReconcileKibanaServices(ctx context.Context, kb *dbv1.Kibana, req ctrl.Request, c client.Client, Scheme *runtime.Scheme) error {
	log := log.FromContext(ctx)
	selector := map[string]string{
		NameLabel:      "es-cluster",
		InstanceLabel:  kb.Name,
		ManagedByLabel: "es-operator",
		PartOfLabel:    "es-cluster",
	}
	svc := &k8scorev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: kb.Namespace,
			Name:      kb.Name,
			Labels:    selector,
		},
		Spec: k8scorev1.ServiceSpec{
			Ports: []k8scorev1.ServicePort{
				{
					Name:     "http",
					Port:     5601,
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
	if err := controllerutil.SetControllerReference(kb, svc, Scheme); err != nil {
		log.Error(err, "SetControllerReference error")
		return err
	}
	// 创建service
	// log.Info("start create service")

	if err := c.Create(ctx, svc); err != nil {
		if errors.IsAlreadyExists(err) {
			if err := c.Update(ctx, svc); err != nil {
				log.Error(err, "create service error")
				return err
			}
		}
	}
	return nil
}
