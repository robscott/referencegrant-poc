package main

import (
	"context"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type ClusterReferencePatternHandler struct {
	c *Controller
}

func NewClusterReferencePatternHandler(c *Controller) *ClusterReferencePatternHandler {
	return &ClusterReferencePatternHandler{c: c}
}

func (h *ClusterReferencePatternHandler) Create(ctx context.Context, e event.CreateEvent, q workqueue.RateLimitingInterface) {
	h.c.log.Info("ClusterReferencePattern Create", e)
	q.Add(reconcile.Request{NamespacedName: types.NamespacedName{Name: e.Object.GetName(), Namespace: e.Object.GetNamespace()}})
}

func (h *ClusterReferencePatternHandler) Update(ctx context.Context, e event.UpdateEvent, q workqueue.RateLimitingInterface) {
	h.c.log.Info("ClusterReferencePattern Update", e)
}

func (h *ClusterReferencePatternHandler) Delete(ctx context.Context, e event.DeleteEvent, q workqueue.RateLimitingInterface) {
	h.c.log.Info("ClusterReferencePattern Delete", e)
}

func (h *ClusterReferencePatternHandler) Generic(ctx context.Context, e event.GenericEvent, q workqueue.RateLimitingInterface) {
	h.c.log.Info("ClusterReferencePattern Generic", e)
}
