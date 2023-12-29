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
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
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
	c.log.Info("Reconciling ClusterReferencePattern", "name", req.NamespacedName.Name)

	// For some very strange reason, CR client expects "default" namespace for
	// cluster-scoped resources and will fail without it being set.
	req.NamespacedName.Namespace = "default"

	crp := &v1a1.ClusterReferencePattern{}
	err := c.crClient.Get(ctx, req.NamespacedName, crp)
	if err != nil {
		if errors.IsNotFound(err) {
			// TODO: Some form of cleanup is needed here
		}
		c.log.Error(err, "error fetching ClusterReferencePattern")
		return ctrl.Result{}, err
	}

	// TODO: Have informers for each target resource of a ClusterReferencePattern
	targetGVR := schema.GroupVersionResource{Group: crp.Group, Version: crp.Version, Resource: crp.Resource}
	targetList, err := c.dClient.Resource(targetGVR).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		c.log.Error(err, "failed to list target for ClusterReferencePattern", targetGVR)
		return ctrl.Result{}, err
	}

	refs := c.getReferences(ctx, targetList, crp.Path)

	crcList := &v1a1.ClusterReferenceConsumerList{}
	err = c.crClient.List(ctx, crcList)
	if err != nil {
		c.log.Error(err, "could not list ClusterReferenceConsumers")
		return ctrl.Result{}, err
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

func (c *Controller) reconcileRBAC(ctx context.Context, crp *v1a1.ClusterReferencePattern, subjects []rbacv1.Subject, references []reference) error {
	var err error
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
	desiredRoles := map[string]rbacv1.Role{}

	for ns, r := range namespaceResourceNames {
		role := rbacv1.Role{
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

	roleList := rbacv1.RoleList{}
	err = c.crClient.List(ctx, &roleList, listOption)
	if err != nil {
		c.log.Error(err, "error listing Roles")
		return err
	}

	// TODO: Add proper reconciliation logic to compare desired and existing so
	// we're not always creating new resources.
	for _, dr := range desiredRoles {
		err := c.crClient.Create(ctx, &dr)
		if err != nil {
			c.log.Error(err, "error creating Role")
			return err
		}
		namespaceRoleNames[dr.Namespace] = dr.Name
	}

	roleBindingList := rbacv1.RoleBindingList{}
	err = c.crClient.List(ctx, &roleBindingList, listOption)
	if err != nil {
		c.log.Error(err, "error listing RoleBindings")
		return err
	}

	for ns, roleName := range namespaceRoleNames {
		rb := rbacv1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: fmt.Sprintf("%s-", crp.Name),
				Namespace:    ns,
				Labels:       map[string]string{labelKeyPatternName: crp.Name},
			},
			Subjects: subjects,
			RoleRef: rbacv1.RoleRef{
				APIGroup: rbacv1.SchemeGroupVersion.Group,
				Kind:     "Role",
				Name:     roleName,
			},
		}
		err := c.crClient.Create(ctx, &rb)
		if err != nil {
			c.log.Error(err, "error creating RoleBinding")
			return err
		}
	}

	c.log.Info("Completed RBAC Reconciliation", "rolesCreated", len(desiredRoles), "roleBindingsCreated", len(namespaceRoleNames))

	return nil
}
