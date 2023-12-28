package main

import (
	"context"
	"os"
	"strings"

	v1a1 "sigs.k8s.io/referencegrant-poc/apis/v1alpha1"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/v2/klogr"
	"k8s.io/klog/v2/textlogger"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Controller struct {
	dClient  *dynamic.DynamicClient
	crClient client.Client
	log      logr.Logger
}

func NewController() *Controller {
	lConfig := textlogger.NewConfig()

	c := &Controller{
		log: textlogger.NewLogger(lConfig),
	}
	ctrl.SetLogger(klogr.New())

	c.log.Info("Initializing Controller")

	kConfig := ctrl.GetConfigOrDie()
	scheme := scheme.Scheme
	v1a1.AddToScheme(scheme)

	dClient, err := dynamic.NewForConfig(kConfig)
	if err != nil {
		c.log.Error(err, "could not create Dynamic client")
		os.Exit(1)
	}

	c.dClient = dClient

	config := ctrl.GetConfigOrDie()
	manager, err := ctrl.NewManager(config, ctrl.Options{Scheme: scheme})
	if err != nil {
		c.log.Error(err, "could not create manager")
		os.Exit(1)
	}

	c.crClient = manager.GetClient()

	err = ctrl.NewControllerManagedBy(manager).
		Named("referencegrant-poc").
		Watches(&v1a1.ClusterReferenceConsumer{}, NewClusterReferenceConsumerHandler(c)).
		Watches(&v1a1.ClusterReferencePattern{}, NewClusterReferencePatternHandler(c)).
		Watches(&v1a1.ReferenceGrant{}, NewReferenceGrantHandler(c)).
		Complete(c)

	if err != nil {
		c.log.Error(err, "could not setup controller")
		os.Exit(1)
	}

	if err := manager.Start(ctrl.SetupSignalHandler()); err != nil {
		c.log.Error(err, "could not start manager")
		os.Exit(1)
	}

	return c
}

func (c *Controller) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	c.log.Info("Reconcile", req.NamespacedName)
	crp := &v1a1.ClusterReferencePattern{}
	err := c.crClient.Get(ctx, req.NamespacedName, crp)
	if err != nil {
		if errors.IsNotFound(err) {
			// TODO
			return ctrl.Result{}, nil
		}
		c.log.Error(err, "could not start manager")
		return ctrl.Result{}, err
	}
	c.log.Info("CRP", crp.Name)
	gvr := schema.GroupVersionResource{Group: crp.Group, Version: crp.Version, Resource: crp.Resource}
	list, err := c.dClient.Resource(gvr).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		c.log.Error(err, "failed to list GVR", gvr)
		return ctrl.Result{}, err
	}

	segments := strings.Split(crp.Path, "[*].")
	c.log.Info("segments", "s", segments, "l", list.IsList())

	ref := make([]interface{}, 0)
	for _, r := range list.Items {
		c.log.Info("Resource", "r", r)

		subSegments := strings.Split(segments[0], ".")
		c.log.Info("subSegments", "s", subSegments)

		matchingSlice, _, _ := unstructured.NestedSlice(r.Object, subSegments...)

		c.log.Info("matchingSlice", "r", matchingSlice)

		for _, match := range matchingSlice {
			subSegments2 := strings.Split(segments[len(segments)-1], ".")
			c.log.Info("match", "r", match, "s", subSegments2)
			ref, _, _ = unstructured.NestedSlice(match.(map[string]interface{}), subSegments2...)
			c.log.Info("Ref", "r", ref)
		}
	}

	return ctrl.Result{}, nil
}

func main() {
	NewController()

	// ReferenceGrant changes
	// - New: Generate new RoleBindings for pattern
	// - Update: Change/diff RoleBindings for pattern
	// - Delete: Update/remove RoleBindings for pattern

	// ClusterReferenceConsumer changes
	// - New: Add subject to all role bindings for pattern
	// - Update: Change subject == change to subject in relevant role bindings
	// - Delete Remove subject from all role bindings

	// ClusterReferencePattern changes
	// - New: Add role bindings for all patterns with empty subject, use predefined label, r+w lock cache for pattern until it's built out
	// - Update: Change role bindings for all patterns, r+w lock cache for pattern until it's built out
	// - Delete: Delete role bindings, r+w lock cache for pattern until it's deleted

	// ClusterReferencePattern Resource changes
	// - New: Create RoleBinding for pattern
	//
	//

	// ReconcilePattern
	// 1) Get all consumers of pattern, derive subjects from that
	// 2) Get current set of role bindings generated for this pattern via label
	// 3) Get desired set of role bindings from cache of pattern references - needs to be rebuilt for some ClusterReferencePattern changes - those should lock cache
	// 4) Update existing role bindings, create missing ones, delete unnecessary
}
