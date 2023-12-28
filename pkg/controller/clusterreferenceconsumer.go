package main

import (
	"context"

	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

type ClusterReferenceConsumerHandler struct {
	c *Controller
}

func NewClusterReferenceConsumerHandler(c *Controller) *ClusterReferenceConsumerHandler {
	return &ClusterReferenceConsumerHandler{c: c}
}

func (h *ClusterReferenceConsumerHandler) Create(ctx context.Context, e event.CreateEvent, q workqueue.RateLimitingInterface) {
	h.c.log.Info("ClusterReferenceConsumer Create", e)
}

func (h *ClusterReferenceConsumerHandler) Update(ctx context.Context, e event.UpdateEvent, q workqueue.RateLimitingInterface) {
	h.c.log.Info("ClusterReferenceConsumer Update", e)
}

func (h *ClusterReferenceConsumerHandler) Delete(ctx context.Context, e event.DeleteEvent, q workqueue.RateLimitingInterface) {
	h.c.log.Info("ClusterReferenceConsumer Delete", e)
}

func (h *ClusterReferenceConsumerHandler) Generic(ctx context.Context, e event.GenericEvent, q workqueue.RateLimitingInterface) {
	h.c.log.Info("ClusterReferenceConsumer Generic", e)
}
