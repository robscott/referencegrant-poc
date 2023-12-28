package main

import (
	"context"

	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

type ReferenceGrantHandler struct {
	c *Controller
}

func NewReferenceGrantHandler(c *Controller) *ReferenceGrantHandler {
	return &ReferenceGrantHandler{c: c}
}

func (h *ReferenceGrantHandler) Create(ctx context.Context, e event.CreateEvent, q workqueue.RateLimitingInterface) {
	h.c.log.Info("ReferenceGrant Create", e)
}

func (h *ReferenceGrantHandler) Update(ctx context.Context, e event.UpdateEvent, q workqueue.RateLimitingInterface) {
	h.c.log.Info("ReferenceGrant Update", e)
}

func (h *ReferenceGrantHandler) Delete(ctx context.Context, e event.DeleteEvent, q workqueue.RateLimitingInterface) {
	h.c.log.Info("ReferenceGrant Delete", e)
}

func (h *ReferenceGrantHandler) Generic(ctx context.Context, e event.GenericEvent, q workqueue.RateLimitingInterface) {
	h.c.log.Info("ReferenceGrant Generic", e)
}
