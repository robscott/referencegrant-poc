package main

import (
	"context"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	v1a1 "sigs.k8s.io/referencegrant-poc/apis/v1alpha1"
)

type ClusterReferenceConsumerHandler struct {
	c *Controller
}

func NewClusterReferenceConsumerHandler(c *Controller) *ClusterReferenceConsumerHandler {
	return &ClusterReferenceConsumerHandler{c: c}
}

func (h *ClusterReferenceConsumerHandler) Create(ctx context.Context, e event.CreateEvent, q workqueue.RateLimitingInterface) {
	queuePatternsForCRC(e.Object, q)
}

func (h *ClusterReferenceConsumerHandler) Update(ctx context.Context, e event.UpdateEvent, q workqueue.RateLimitingInterface) {
	queuePatternsForCRC(e.ObjectNew, q)
	queuePatternsForCRC(e.ObjectOld, q)
}

func (h *ClusterReferenceConsumerHandler) Delete(ctx context.Context, e event.DeleteEvent, q workqueue.RateLimitingInterface) {
	queuePatternsForCRC(e.Object, q)
}

func (h *ClusterReferenceConsumerHandler) Generic(ctx context.Context, e event.GenericEvent, q workqueue.RateLimitingInterface) {
	queuePatternsForCRC(e.Object, q)
}

func queuePatternsForCRC(obj client.Object, q workqueue.RateLimitingInterface) {
	crc := obj.(*v1a1.ClusterReferenceConsumer)
	for _, pn := range crc.PatternNames {
		q.AddRateLimited(reconcile.Request{NamespacedName: types.NamespacedName{Name: pn}})
	}
}
