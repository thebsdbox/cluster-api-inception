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
	"fmt"

	"github.com/go-logr/logr"
	"github.com/prometheus/common/log"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha2"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	infrav1 "github.com/thebsdbox/cluster-api-inception/api/v1alpha1"
)

// InceptionMachineReconciler reconciles a InceptionMachine object
type InceptionMachineReconciler struct {
	client.Client
	Log logr.Logger
}

// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=inceptionmachines,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=inceptionmachines/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=clusters;machines;machines/status,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=events;secrets,verbs=get;list;watch;create;update;patch

func (r *InceptionMachineReconciler) Reconcile(req ctrl.Request) (_ ctrl.Result, reterr error) {
	ctx := context.Background()
	logger := r.Log.WithValues("inceptionmachine", req.NamespacedName)

	// This is where we start our controller logic

	// Fetch the inceptionmachine instance.
	inceptionMachine := &infrav1.InceptionMachine{}

	err := r.Get(ctx, req.NamespacedName, inceptionMachine)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	logger = logger.WithName(inceptionMachine.APIVersion)

	// Fetch the Machine.
	machine, err := util.GetOwnerMachine(ctx, r.Client, inceptionMachine.ObjectMeta)
	if err != nil {
		return reconcile.Result{}, err
	}
	if machine == nil {
		logger.Info("Machine Controller has not yet set OwnerRef")
		return reconcile.Result{}, nil
	}
	logger = logger.WithName(fmt.Sprintf("machine=%s", machine.Name))

	// Fetch the Cluster.
	cluster, err := util.GetClusterFromMetadata(ctx, r.Client, machine.ObjectMeta)
	if err != nil {
		logger.Info("Machine is missing cluster label or cluster does not exist")
		return reconcile.Result{}, nil
	}

	logger = logger.WithName(fmt.Sprintf("cluster=%s", cluster.Name))

	// Fetch the inception Cluster
	inceptionCluster := &infrav1.InceptionCluster{}
	inceptionClusterName := types.NamespacedName{
		Namespace: inceptionMachine.Namespace, // get the name from the machine
		Name:      cluster.Spec.InfrastructureRef.Name,
	}

	if err := r.Client.Get(ctx, inceptionClusterName, inceptionCluster); err != nil {
		logger.Info("The Inception Cluster is not available yet")
		return reconcile.Result{}, nil
	}

	logger = logger.WithName(fmt.Sprintf("inceptionCluster=%s", inceptionCluster.Name))

	// Initialize the patch helper
	patchHelper, err := patch.NewHelper(inceptionMachine, r)
	if err != nil {
		return ctrl.Result{}, err
	}
	// Always attempt to Patch the Machine object and status after each reconciliation.
	defer func() {
		if err := patchHelper.Patch(ctx, inceptionMachine); err != nil {
			if reterr == nil {
				reterr = err
			}
		}
	}()
	// Handle deleted clusters
	if !inceptionMachine.DeletionTimestamp.IsZero() {
		return r.reconcileMachineDelete(logger, machine, inceptionMachine, cluster, inceptionCluster)
	}

	// Handle non-deleted clusters
	return r.reconcileMachine(logger, machine, inceptionMachine, cluster, inceptionCluster)
}

func (r *InceptionMachineReconciler) reconcileMachine(logger logr.Logger, machine *clusterv1.Machine, inceptionMachine *infrav1.InceptionMachine, cluster *clusterv1.Cluster, inceptionCluster *infrav1.InceptionCluster) (_ ctrl.Result, reterr error) {
	logger.Info("Reconciling Machine")
	// If the DockerMachine doesn't have finalizer, add it.
	if !util.Contains(inceptionMachine.Finalizers, infrav1.MachineFinalizer) {
		inceptionMachine.Finalizers = append(inceptionMachine.Finalizers, infrav1.MachineFinalizer)
	}

	// if the machine is already provisioned, return
	if inceptionMachine.Spec.ProviderID != nil {
		inceptionMachine.Status.Ready = true

		return ctrl.Result{}, nil
	}

	// Make sure bootstrap data is available and populated.
	if machine.Spec.Bootstrap.Data == nil {

		log.Info("Waiting for the Bootstrap provider controller to set bootstrap data")
		//return ctrl.Result{}, nil
	}

	// Check the role of the machine
	// role := constants.WorkerNodeRoleValue
	// if util.IsControlPlaneMachine(inceptionMachine) {
	// 	role = constants.ControlPlaneNodeRoleValue
	// }

	// TODO - Attempt to create the machine

	// // if the machine is a control plane added, update the load balancer configuration
	// if util.IsControlPlaneMachine(machine) {}

	providerID := "inception:////inception"
	inceptionMachine.Spec.ProviderID = &providerID
	// Mark the inceptionMachine ready
	inceptionMachine.Status.Ready = true

	return ctrl.Result{}, nil

}

func (r *InceptionMachineReconciler) reconcileMachineDelete(logger logr.Logger, machine *clusterv1.Machine, inceptionMachine *infrav1.InceptionMachine, cluster *clusterv1.Cluster, inceptionCluster *infrav1.InceptionCluster) (_ ctrl.Result, reterr error) {
	logger.Info("Deleting Machine")
	// Machine is deleted so remove the finalizer.
	inceptionMachine.Finalizers = util.Filter(inceptionMachine.Finalizers, infrav1.MachineFinalizer)
	return ctrl.Result{}, nil

}

func (r *InceptionMachineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrav1.InceptionMachine{}).Watches(
		&source.Kind{Type: &clusterv1.Machine{}},
		&handler.EnqueueRequestsFromMapFunc{
			ToRequests: util.MachineToInfrastructureMapFunc(infrav1.GroupVersion.WithKind("InceptionMachine")),
		},
	).
		Complete(r)
}
