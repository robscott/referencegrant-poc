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

	v1a1 "sigs.k8s.io/referencegrant-poc/apis/v1alpha1"

	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
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
	nn := generateQueueKey(crc.From.Group, crc.From.Resource, crc.To.Group, crc.To.Resource, string(crc.For))
	q.AddRateLimited(reconcile.Request{nn})
}
