package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/kubernetes-sigs/ingress2gateway/pkg/i2gw"
	"github.com/kubernetes-sigs/ingress2gateway/pkg/i2gw/providers/common"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

type IngressReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

func (r *IngressReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	log := r.Log.WithValues("ingress", req.NamespacedName)
	log.Info("Reconciling ingress creation request")

	instance := &networkingv1.Ingress{}
	err = r.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Ingress not found. Ignoring since object must be deleted")
			return ctrl.Result{RequeueAfter: time.Second * 5}, nil
		}
		log.Error(err, "Failed to get Ingress object")
		return ctrl.Result{}, err
	}

	resources, errs := common.ToGateway([]networkingv1.Ingress{*instance}, i2gw.ProviderImplementationSpecificOptions{})
	if len(errs) > 0 {
		for _, e := range errs {
			log.Error(e, "Failed to convert Ingress to Gateway")
		}
		return ctrl.Result{}, errs[0]
	}

	// for _, v := range resources.Gateways {
	// 	err = createOrUpdateGateway(ctx, &v, r.Client)
	// 	if err != nil {
	// 		log.Error(err, "Failed to create or update Gateway")
	// 		return ctrl.Result{}, err
	// 	}
	// }

	for _, v := range resources.HTTPRoutes {
		err = createOrUpdateHttpRoute(ctx, &v, r.Client)
		if err != nil {
			log.Error(err, "Failed to create or update HTTPRoute")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func createOrUpdateHttpRoute(ctx context.Context, desired *gwapiv1.HTTPRoute, c client.Client) error {
	existing := desired.DeepCopy()
	_, err := controllerutil.CreateOrUpdate(ctx, c, existing, func() error {
		existing.Labels = desired.Labels
		existing.Annotations = mergeMaps(desired.Annotations, existing.Annotations)
		existing.OwnerReferences = desired.OwnerReferences
		existing.Spec = desired.Spec
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to create or update HTTPRoute: %w", err)
	}
	existing.DeepCopyInto(desired)
	return nil
}

func createOrUpdateGateway(ctx context.Context, desired *gwapiv1.Gateway, c client.Client) error {
	existing := desired.DeepCopy()
	_, err := controllerutil.CreateOrUpdate(ctx, c, existing, func() error {
		existing.Labels = desired.Labels
		existing.Annotations = mergeMaps(desired.Annotations, existing.Annotations)
		existing.OwnerReferences = desired.OwnerReferences
		existing.Spec = desired.Spec
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to create or update HTTPRoute: %w", err)
	}
	existing.DeepCopyInto(desired)
	return nil
}

func (r *IngressReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&networkingv1.Ingress{}).
		Owns(&gwapiv1.HTTPRoute{}).
		Complete(r)
}
