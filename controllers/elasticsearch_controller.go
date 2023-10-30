/*
 * @FilePath: /Users/zengshenglong/Code/GoWorkSpace/operators/es-operator/controllers/elasticsearch_controller.go
 * @Author: kbsonlong kbsonlong@gmail.com
 * @Date: 2023-10-09 13:00:45
 * @LastEditors: kbsonlong kbsonlong@gmail.com
 * @LastEditTime: 2023-10-25 17:09:23
 * @Description:
 * Copyright (c) 2023 by kbsonlong, All Rights Reserved.
 */
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
	"encoding/json"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	elasticsearch7 "github.com/elastic/go-elasticsearch/v7"
	dbv1 "github.com/kbsonlong/es-operator/api/v1"
	"github.com/kbsonlong/es-operator/pkg/k8s"
	apps "k8s.io/api/apps/v1"
	k8scorev1 "k8s.io/api/core/v1"
)

// ElasticsearchReconciler reconciles a Elasticsearch object
type ElasticsearchReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

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

	es := &dbv1.Elasticsearch{}
	// fmt.Println(elastic.BaseConfig(es.Name, "ipv4"))
	// cfg := elastic.BaseConfig(es.Name, "ipv4")
	// fmt.Println(cfg)
	// for _, nodeSpec := range es.Spec.NodeSets {
	// 	// build es config
	// 	userCfg := dbv1.Config{}
	// 	if nodeSpec.Config != nil {
	// 		userCfg = *nodeSpec.Config
	// 	}
	// 	cfg, err := settings.NewMergedESConfig(es.Name, ver, ipFamily, es.Spec.HTTP, userCfg)
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	// build stateful set and associated headless service
	// 	statefulSet, err := BuildStatefulSet(ctx, client, es, nodeSpec, cfg, keystoreResources, existingStatefulSets, setDefaultSecurityContext)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	headlessSvc := HeadlessService(&es, statefulSet.Name)

	// 	nodesResources = append(nodesResources, Resources{
	// 		NodeSet:         nodeSpec.Name,
	// 		StatefulSet:     statefulSet,
	// 		HeadlessService: headlessSvc,
	// 		Config:          cfg,
	// 	})
	// }

	if err := r.Get(ctx, req.NamespacedName, es); err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Return and don't requeue
			logger.Info("ES Cluster resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get ES Cluster")
		return ctrl.Result{RequeueAfter: time.Second * 5}, err
	}

	es.Status.Health = dbv1.ElasticsearchUnknownHealth
	es.Status.Phase = dbv1.ElasticsearchResourceInvalid
	defer func() {
		err := r.updateStatus(ctx, es)
		// err := r.Client.Status().Update(ctx, es)
		if err != nil {
			logger.Error(err, "failed to update cluster status")
		}
	}()

	if err := k8s.ReconcileConfigMap(ctx, es, req, r.Client, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}
	if err := k8s.ReconcileElasticServices(ctx, es, req, r.Client, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}
	if err := k8s.ReconcileStatefulSet(ctx, es, req, r.Client, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	err := r.updateStatus(ctx, es)

	return ctrl.Result{}, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *ElasticsearchReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dbv1.Elasticsearch{}).
		Owns(&apps.StatefulSet{}).
		Owns(&k8scorev1.Service{}).
		Owns(&k8scorev1.ConfigMap{}).
		Complete(r)
}

func GetHealth(es *dbv1.Elasticsearch) (map[string]interface{}, error) {
	cfg := elasticsearch7.Config{
		Addresses: []string{
			fmt.Sprintf("http://%s.%s.svc.%s:9200", es.Name, es.Namespace, dbv1.DefaultDomain),
		},
		// Base Authentication
		// Username: "elastic",
		// Password: "5Z1YI056Zc9RtQt51nBrm5p7",
		// Disable Tls
		// Transport: &http.Transport{
		// 	TLSClientConfig: &tls.Config{
		// 		InsecureSkipVerify: true,
		// 	},
		// },
	}
	esClient, err := elasticsearch7.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	var data map[string]interface{}

	response, err := esClient.Cluster.Health()
	if err != nil {
		return nil, err
	}

	buf.ReadFrom(response.Body)
	respBytes := buf.String()
	respString := string(respBytes)

	err = json.Unmarshal([]byte(respString), &data)
	if err != nil {
		fmt.Println("json data", err)
	}

	return data, err
}

func (r *ElasticsearchReconciler) updateStatus(ctx context.Context, es *dbv1.Elasticsearch) error {
	logger := log.FromContext(ctx)

	data, err := GetHealth(es)
	if err != nil {
		r.Status().Update(ctx, es)
		return err
	}

	availableNodes, _ := data["number_of_nodes"].(float64)
	availableDataNodes, _ := data["number_of_data_nodes"].(float64)
	status, _ := data["status"].(string)

	es.Status.AvailableNodes = int32(availableNodes)
	es.Status.AvailableDataNodes = int32(availableDataNodes)
	es.Status.Health = dbv1.ElasticsearchHealth(status)
	logger.Info("Update Elasticsearch Status")
	pods, err := k8s.GetPods(ctx, es, r.Client)

	if len(pods.Items) != int(es.Spec.Size) {
		es.Status.Phase = dbv1.ElasticsearchApplyingChangesPhase
		r.Status().Update(ctx, es)
		err = fmt.Errorf("waiting pods %d to %d ", len(pods.Items), es.Spec.Size)
		return err
	} else if es.Status.AvailableNodes != es.Spec.Size {
		es.Status.Phase = dbv1.ElasticsearchApplyingChangesPhase
		r.Status().Update(ctx, es)
		err = fmt.Errorf("waiting for elasticsearch available nodes %d to %d ", es.Status.AvailableNodes, es.Spec.Size)
		return err
	} else if es.Status.Health != "green" {
		es.Status.Phase = dbv1.ElasticsearchApplyingChangesPhase
		r.Status().Update(ctx, es)
		err = fmt.Errorf("waiting for elasticsearch  cluster status are green")
		return err
	} else {
		es.Status.Phase = dbv1.ElasticsearchReadyPhase
		fmt.Println("Update Elasticsearch Status ready")
		r.Status().Update(ctx, es)
	}

	return err
}
