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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	v1a1 "sigs.k8s.io/referencegrant-poc/apis/v1alpha1"

	"github.com/go-logr/logr"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/util/jsonpath"
	"k8s.io/klog/v2/klogr"
	"k8s.io/klog/v2/textlogger"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	labelKeyPatternName = "reference.authorization.k8s.io/pattern-name"
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

	manager, err := ctrl.NewManager(kConfig, ctrl.Options{Scheme: scheme})
	if err != nil {
		c.log.Error(err, "could not create manager")
		os.Exit(1)
	}

	c.crClient = manager.GetClient()

	// TODO: Add selective ClusterRole and RoleBinding watchers here
	err = ctrl.NewControllerManagedBy(manager).
		Named("referencegrant-poc").
		Watches(&v1a1.ClusterReferenceConsumer{}, NewClusterReferenceConsumerHandler(c)).
		Watches(&v1a1.ClusterReferenceGrant{}, NewClusterReferenceGrantHandler(c)).
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
	c.log.Info("Reconciling for", "name", req.NamespacedName.Name)

	// fromGroup, fromResource, toGroup, toResource, forReason := parseQueueKey(req.NamespacedName)
	referenceGrants, clusterReferenceGrants, clusterReferenceConsumers, err := c.getResourcesFor(ctx, req.NamespacedName.Name)

	for _, crg := range clusterReferenceGrants {
		// TODO: Have informers for each target resource of a ClusterReferenceGrant
		targetGVR := schema.GroupVersionResource{Group: crg.From.Group, Version: crg.From.Version, Resource: crg.From.Version}
		targetList, err := c.dClient.Resource(targetGVR).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			c.log.Error(err, "failed to list target for ClusterReferenceGrant", targetGVR)
			return ctrl.Result{}, err
		}
		refs := c.getReferences(ctx, targetList, crc.Path)
	}

	subjects := c.getSubjects(ctx, crcList, crp.Name)

	// TODO: Don't just blindly trust references, ensure ReferenceGrant allows them first
	err = c.reconcileRBAC(ctx, crp, subjects, refs)
	if err != nil {
		c.log.Error(err, "error reconciling RBAC")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (c *Controller) getSubjects(ctx context.Context, list *v1a1.ClusterReferenceConsumerList, patternName string) []rbacv1.Subject {
	subjects := []rbacv1.Subject{}

	for _, crc := range list.Items {
		match := false
		for _, pn := range crc.PatternNames {
			if pn == patternName {
				match = true
				break
			}
		}

		if match {
			// TODO: Dedupe
			subjects = append(subjects, crc.Subject)
		}
	}

	return subjects
}

type reference struct {
	Group         string
	Resource      string
	FromNamespace string
	ToNamespace   string
	Name          string
}

func (c *Controller) getReferences(ctx context.Context, list *unstructured.UnstructuredList, path string) []reference {
	refs := []reference{}
	for _, item := range list.Items {
		j := jsonpath.New("test")
		err := j.Parse(fmt.Sprintf("{%s}", path))
		if err != nil {
			c.log.Error(err, "error parsing JSON Path")
		}
		results := new(bytes.Buffer)
		err = j.Execute(results, item.UnstructuredContent())
		if err != nil {
			c.log.Error(err, "error finding results with JSON Path")
		}

		rawRefs := strings.Split(results.String(), " ")

		for _, rr := range rawRefs {
			jr := map[string]string{}
			err = json.Unmarshal([]byte(rr), &jr)
			group, hasGroup := jr["group"]
			if !hasGroup {
				c.log.Info("Missing group in reference", "ref", jr)
				continue
			}
			resource, hasResource := jr["resource"]
			if !hasResource {
				kind, hasKind := jr["kind"]
				if !hasKind {
					c.log.Info("Missing kind or resource in reference", "ref", jr)
					continue
				}
				gvr, _ := meta.UnsafeGuessKindToResource(schema.GroupVersionKind{Group: group, Version: "v1", Kind: kind})
				resource = gvr.Resource
			}

			namespace, hasNamespace := jr["namespace"]
			if !hasNamespace {
				namespace = item.GetNamespace()
			}

			name, hasName := jr["name"]
			if !hasName {
				c.log.Info("Missing name in reference", "ref", jr)
				continue
			}

			refs = append(refs, reference{
				Group:         group,
				Resource:      resource,
				FromNamespace: item.GetNamespace(),
				ToNamespace:   namespace,
				Name:          name,
			})
		}
	}

	return refs
}

// Format: group/resource
type groupResource string
type resourceNamesByGroupAndResource map[groupResource]sets.Set[string]

func (r *reference) GroupResource() groupResource {
	return groupResource(fmt.Sprintf("%s/%s", r.Group, r.Resource))
}

// TODO: This is awful, find a better approach
func splitGroupResource(gr groupResource) (string, string) {
	s := strings.Split(string(gr), "/")
	return s[0], s[1]
}

type reconciliationResults struct {
	rolesCreated        uint
	rolesUpdated        uint
	rolesDeleted        uint
	roleBindingsCreated uint
	roleBindingsUpdated uint
	roleBindingsDeleted uint
}

func (c *Controller) reconcileRBAC(ctx context.Context, crp *v1a1.ClusterReferenceGrant, subjects []rbacv1.Subject, references []reference) error {
	var err error
	rr := reconciliationResults{}
	listOption := client.MatchingLabels{
		labelKeyPatternName: crp.Name,
	}

	// TODO: Clean this up + extract it out
	// Namespace -> Group+Resource -> Resource Name
	namespaceResourceNames := map[string]resourceNamesByGroupAndResource{}
	for _, ref := range references {
		r, hasNamespace := namespaceResourceNames[ref.ToNamespace]
		if !hasNamespace {
			r = resourceNamesByGroupAndResource{}
			namespaceResourceNames[ref.ToNamespace] = r
		}
		gr := ref.GroupResource()
		names, hasNames := r[gr]
		if !hasNames {
			names = sets.New[string]()
			r[gr] = names
		}
		names.Insert(ref.Name)
	}

	baseVerbs := []string{"get", "watch", "list"}
	namespaceRoleNames := map[string]string{}
	desiredRoles := map[string]*rbacv1.Role{}

	for ns, r := range namespaceResourceNames {
		role := &rbacv1.Role{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: fmt.Sprintf("%s-", crp.Name),
				Namespace:    ns,
				Labels:       map[string]string{labelKeyPatternName: crp.Name},
			},
		}
		for gr, nameSet := range r {
			group, resource := splitGroupResource(gr)
			names := nameSet.UnsortedList()
			role.Rules = append(role.Rules, rbacv1.PolicyRule{
				APIGroups:     []string{group},
				Resources:     []string{resource},
				Verbs:         baseVerbs,
				ResourceNames: names,
			})
		}
		desiredRoles[ns] = role
	}

	existingRoles := map[string]rbacv1.Role{}
	roleList := rbacv1.RoleList{}
	rolesToDelete := []rbacv1.Role{}
	err = c.crClient.List(ctx, &roleList, listOption)
	if err != nil {
		c.log.Error(err, "error listing Roles")
		return err
	}
	for _, role := range roleList.Items {
		existingRole, isExisting := existingRoles[role.Namespace]
		desiredRole, isDesired := desiredRoles[role.Namespace]

		// We want at most one role per ClusterReferenceGrant and Namespace,
		// anything beyond that should be deleted.
		if !isExisting && isDesired {
			existingRole = role
			existingRoles[role.Namespace] = existingRole
			desiredRole.Name = existingRole.Name
			desiredRole.GenerateName = ""
		} else {
			rolesToDelete = append(rolesToDelete, role)
		}
	}

	// TODO: Add proper reconciliation logic to compare desired and existing so
	// we're not always updating resources even if they don't need to change.
	for _, dr := range desiredRoles {
		if dr.Name != "" {
			c.log.Info("Updating role", "role", dr)
			err := c.crClient.Update(ctx, dr)
			if err != nil {
				c.log.Error(err, "error updating Role")
				return err
			}
			rr.rolesUpdated++
		} else {
			c.log.Info("Creating role", "role", dr)
			err := c.crClient.Create(ctx, dr)
			if err != nil {
				c.log.Error(err, "error creating Role")
				return err
			}
			rr.rolesCreated++
		}
		namespaceRoleNames[dr.Namespace] = dr.Name
	}

	for _, rtd := range rolesToDelete {
		c.log.Info("Deleting role", "role", rtd)
		err := c.crClient.Delete(ctx, &rtd)
		if err != nil {
			c.log.Error(err, "error deleting Role")
			return err
		}
		rr.rolesDeleted++
	}

	existingRoleBindings := map[string]rbacv1.RoleBinding{}
	roleBindingList := rbacv1.RoleBindingList{}
	roleBindingsToDelete := []rbacv1.RoleBinding{}
	err = c.crClient.List(ctx, &roleBindingList, listOption)
	if err != nil {
		c.log.Error(err, "error listing RoleBindings")
		return err
	}

	for _, rb := range roleBindingList.Items {
		_, isExisting := existingRoleBindings[rb.Namespace]
		desiredRole, isDesired := desiredRoles[rb.Namespace]

		// We want at most one RoleBinding per ClusterReferenceGrant and
		// Namespace, anything beyond that should be deleted. We also can't
		// change the RoleRef on an existing RoleBinding.
		if !isExisting && isDesired && rb.RoleRef.Name == desiredRole.Name {
			existingRoleBindings[rb.Namespace] = rb
		} else {
			roleBindingsToDelete = append(roleBindingsToDelete, rb)
		}
	}

	for ns, roleName := range namespaceRoleNames {
		rb := rbacv1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns,
				Labels:    map[string]string{labelKeyPatternName: crp.Name},
			},
			Subjects: subjects,
			RoleRef: rbacv1.RoleRef{
				APIGroup: rbacv1.SchemeGroupVersion.Group,
				Kind:     "Role",
				Name:     roleName,
			},
		}
		if existingRB, ok := existingRoleBindings[rb.Namespace]; ok {
			c.log.Info("Updating RoleBinding", "RoleBinding", rb)
			rb.Name = existingRB.Name
			err := c.crClient.Update(ctx, &rb)
			if err != nil {
				c.log.Error(err, "error updating RoleBinding")
				return err
			}
			rr.roleBindingsUpdated++
		} else {
			rb.GenerateName = fmt.Sprintf("%s-", crp.Name)
			c.log.Info("Creating RoleBinding", "RoleBinding", rb)
			err := c.crClient.Create(ctx, &rb)
			if err != nil {
				c.log.Error(err, "error creating RoleBinding")
				return err
			}
			rr.roleBindingsCreated++
		}
	}

	for _, rbtd := range roleBindingsToDelete {
		c.log.Info("Deleting RoleBinding", "RoleBinding", rbtd)
		err := c.crClient.Delete(ctx, &rbtd)
		if err != nil {
			c.log.Error(err, "error deleting RoleBinding")
			return err
		}
		rr.roleBindingsDeleted++
	}

	c.log.Info("Completed RBAC Reconciliation", "Results", fmt.Sprintf("%+v", rr))

	return nil
}
func (c *Controller) getResourcesFor(ctx context.Context, forReason string) ([]v1a1.ReferenceGrant, []v1a1.ClusterReferenceGrant, []v1a1.ClusterReferenceConsumer, error) {
	rgList := &v1a1.ReferenceGrantList{}
	err := c.crClient.List(ctx, rgList)
	if err != nil {
		c.log.Error(err, "could not list ReferenceGrants")
		return nil, nil, nil, err
	}
	referenceGrants := []v1a1.ReferenceGrant{}
	for _, rg := range rgList.Items {
		if string(rg.For) == forReason {
			referenceGrants = append(referenceGrants, rg)
		}
	}

	crgList := &v1a1.ClusterReferenceGrantList{}
	err = c.crClient.List(ctx, rgList)
	if err != nil {
		c.log.Error(err, "could not list ClusterReferenceGrants")
		return nil, nil, nil, err
	}
	clusterReferenceGrants := []v1a1.ClusterReferenceGrant{}
	for _, crg := range crgList.Items {
		if string(crg.For) == forReason {
			clusterReferenceGrants = append(clusterReferenceGrants, crg)
		}
	}

	crcList := &v1a1.ClusterReferenceConsumerList{}
	err = c.crClient.List(ctx, crcList)
	if err != nil {
		c.log.Error(err, "could not list ClusterReferenceConsumers")
		return nil, nil, nil, err
	}
	clusterReferenceConsumers := []v1a1.ClusterReferenceConsumer{}
	for _, crc := range crcList.Items {
		if string(crc.For) == forReason {
			clusterReferenceConsumers = append(clusterReferenceConsumers, crc)
		}
	}

	return rgList.Items, crgList.Items, crcList.Items, nil
}

func generateQueueKey(fromGroup, fromResource, toGroup, toResource, forReason string) types.NamespacedName {
	nn := types.NamespacedName{Name: forReason}
	nn.Namespace = fmt.Sprintf("%s/%s-%s/%s")
	return nn
}

func parseQueueKey(nn types.NamespacedName) (string, string, string, string, string) {
	fromTo := strings.Split(nn.Namespace, "-")
	from := fromTo[0]
	fromGR := strings.Split(from, "/")
	fromGroup := fromGR[0]
	fromResource := fromGR[1]
	to := fromTo[1]
	toGR := strings.Split(to, "/")
	toGroup := toGR[0]
	toResource := toGR[1]

	return fromGroup, fromResource, toGroup, toResource, nn.Name
}
