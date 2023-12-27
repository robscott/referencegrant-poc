package main

import (
	"context"
	"fmt"
	"os"

	v1a1 "sigs.k8s.io/referencegrant-poc/apis/v1alpha1"

	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func main() {
	var log = ctrl.Log.WithName("referencegrant-poc")
	scheme := scheme.Scheme
	v1a1.AddToScheme(scheme)

	manager, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{})
	if err != nil {
		log.Error(err, "could not create manager")
		os.Exit(1)
	}

	fmt.Printf("Manager =====> %+v\n", manager)

	err = ctrl.
		NewControllerManagedBy(manager).
		For(&v1a1.ReferenceGrant{}).
		Complete(&ReferenceGrantReconciler{Client: manager.GetClient()})
	if err != nil {
		fmt.Printf("Could not create =====> %+v\n", err)
		log.Error(err, "could not create controller")
		os.Exit(1)
	}

	if err := manager.Start(ctrl.SetupSignalHandler()); err != nil {
		log.Error(err, "could not start manager")
		os.Exit(1)
	}
}

// ReferenceGrantReconciler is a simple Controller example implementation.
type ReferenceGrantReconciler struct {
	client.Client
}

func (a *ReferenceGrantReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	fmt.Printf("Reconciling =====> %+v\n", req)
	rg := &v1a1.ReferenceGrant{}
	err := a.Get(ctx, req.NamespacedName, rg)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}
