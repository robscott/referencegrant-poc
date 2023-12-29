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

type ReferenceGrantHandler struct {
	c *Controller
}

func NewReferenceGrantHandler(c *Controller) *ReferenceGrantHandler {
	return &ReferenceGrantHandler{c: c}
}

func (h *ReferenceGrantHandler) Create(ctx context.Context, e event.CreateEvent, q workqueue.RateLimitingInterface) {
	queuePatternForRG(e.Object, q)
}

func (h *ReferenceGrantHandler) Update(ctx context.Context, e event.UpdateEvent, q workqueue.RateLimitingInterface) {
	queuePatternForRG(e.ObjectNew, q)
	queuePatternForRG(e.ObjectOld, q)
}

func (h *ReferenceGrantHandler) Delete(ctx context.Context, e event.DeleteEvent, q workqueue.RateLimitingInterface) {
	queuePatternForRG(e.Object, q)
}

func (h *ReferenceGrantHandler) Generic(ctx context.Context, e event.GenericEvent, q workqueue.RateLimitingInterface) {
	queuePatternForRG(e.Object, q)
}

func queuePatternForRG(obj client.Object, q workqueue.RateLimitingInterface) {
	rg := obj.(*v1a1.ReferenceGrant)
	q.AddRateLimited(reconcile.Request{NamespacedName: types.NamespacedName{Name: rg.PatternName}})
}
