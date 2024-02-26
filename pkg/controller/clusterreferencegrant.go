/*
Copyright 2024 The Kubernetes Authors.

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

package main

import (
	"context"

	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	v1a1 "sigs.k8s.io/referencegrant-poc/apis/v1alpha1"
)

type ClusterReferenceGrantHandler struct {
	c *Controller
}

func NewClusterReferenceGrantHandler(c *Controller) *ClusterReferenceGrantHandler {
	return &ClusterReferenceGrantHandler{c: c}
}

func (h *ClusterReferenceGrantHandler) Create(ctx context.Context, e event.CreateEvent, q workqueue.RateLimitingInterface) {
	queueCRP(e.Object, q)
}

func (h *ClusterReferenceGrantHandler) Update(ctx context.Context, e event.UpdateEvent, q workqueue.RateLimitingInterface) {
	queueCRP(e.ObjectNew, q)
}

func (h *ClusterReferenceGrantHandler) Delete(ctx context.Context, e event.DeleteEvent, q workqueue.RateLimitingInterface) {
	queueCRP(e.Object, q)
}

func (h *ClusterReferenceGrantHandler) Generic(ctx context.Context, e event.GenericEvent, q workqueue.RateLimitingInterface) {
	queueCRP(e.Object, q)
}

func queueCRP(obj client.Object, q workqueue.RateLimitingInterface) {
	crg := obj.(*v1a1.ClusterReferenceGrant)

	nn := generateQueueKey(crg.From.Group, crg.From.Resource, crg.To.Group, crg.To.Resource, string(crg.For))
	q.AddRateLimited(reconcile.Request{nn})
}
