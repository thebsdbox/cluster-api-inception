/*

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
	"context"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha2"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	infrav1 "github.com/thebsdbox/cluster-api-inception/api/v1alpha1"
)

const (
	clusterControllerName = "inceptioncluster-controller"
)

// InceptionClusterReconciler reconciles a InceptionCluster object
type InceptionClusterReconciler struct {
	client.Client
	Log logr.Logger
}

// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=inceptionclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=inceptionclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=clusters;clusters/status,verbs=get;list;watch

func (r *InceptionClusterReconciler) Reconcile(req ctrl.Request) (_ ctrl.Result, rerr error) {
	ctx := context.Background()
	log := log.Log.WithName(clusterControllerName).WithValues("inception-cluster", req.NamespacedName)

	// Fetch the InceptionCluster instance
	inceptionCluster := &infrav1.InceptionCluster{}
	if err := r.Client.Get(ctx, req.NamespacedName, inceptionCluster); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Fetch the Cluster.
	cluster, err := util.GetOwnerCluster(ctx, r.Client, inceptionCluster.ObjectMeta)
	if err != nil {
		return ctrl.Result{}, err
	}
	if cluster == nil {
		log.Info("Waiting for Cluster Controller to set OwnerRef on Plunder Cluster")
		return ctrl.Result{}, nil
	}

	log = log.WithValues("cluster", cluster.Name)

	// Initialize the patch helper
	patchHelper, err := patch.NewHelper(inceptionCluster, r)
	if err != nil {
		return ctrl.Result{}, err
	}
	// Always attempt to Patch the inceptionCluster object and status after each reconciliation.
	defer func() {
		if err := patchHelper.Patch(ctx, inceptionCluster); err != nil {
			log.Error(err, "failed to patch infrav1Cluster")
			if rerr == nil {
				rerr = err
			}
		}
	}()

	// Handle deleted clusters
	if !inceptionCluster.DeletionTimestamp.IsZero() {
		return r.reconcileClusterDelete(log, inceptionCluster)
	}

	return r.reconcileCluster(log, cluster, inceptionCluster)
}

func (r *InceptionClusterReconciler) reconcileCluster(logger logr.Logger, cluster *clusterv1.Cluster, plunderCluster *infrav1.InceptionCluster) (_ ctrl.Result, reterr error) {
	logger.Info("Reconciling Cluster")
	if !util.Contains(plunderCluster.Finalizers, infrav1.ClusterFinalizer) {
		plunderCluster.Finalizers = append(plunderCluster.Finalizers, infrav1.ClusterFinalizer)
	}
	plunderCluster.Status.Ready = true
	return ctrl.Result{}, reterr

}

func (r *InceptionClusterReconciler) reconcileClusterDelete(logger logr.Logger, plunderCluster *infrav1.InceptionCluster) (_ ctrl.Result, reterr error) {
	logger.Info("Deleting Cluster")
	plunderCluster.Finalizers = util.Filter(plunderCluster.Finalizers, infrav1.ClusterFinalizer)

	return ctrl.Result{}, reterr
}

func (r *InceptionClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrav1.InceptionCluster{}).
		Complete(r)
}
