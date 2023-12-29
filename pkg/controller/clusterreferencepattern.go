package main

import (
	"context"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	queueCRP(e.Object, q)
}

func (h *ClusterReferencePatternHandler) Update(ctx context.Context, e event.UpdateEvent, q workqueue.RateLimitingInterface) {
	queueCRP(e.ObjectNew, q)
}

func (h *ClusterReferencePatternHandler) Delete(ctx context.Context, e event.DeleteEvent, q workqueue.RateLimitingInterface) {
	queueCRP(e.Object, q)
}

func (h *ClusterReferencePatternHandler) Generic(ctx context.Context, e event.GenericEvent, q workqueue.RateLimitingInterface) {
	queueCRP(e.Object, q)
}

func queueCRP(obj client.Object, q workqueue.RateLimitingInterface) {
	q.Add(reconcile.Request{NamespacedName: types.NamespacedName{Name: obj.GetName()}})
}
