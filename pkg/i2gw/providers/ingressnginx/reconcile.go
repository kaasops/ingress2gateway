package ingressnginx

import (
	"context"
	"fmt"
	"time"

	"github.com/kubernetes-sigs/ingress2gateway/pkg/i2gw/providers/common"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func (p *Provider) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	log := p.Log.WithValues("ingress", req.NamespacedName)
	log.Info("Reconciling ingress creation request")

	instance := &networkingv1.Ingress{}
	err = p.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Ingress not found. Ignoring since object must be deleted")
			return ctrl.Result{RequeueAfter: time.Second * 5}, nil
		}
		log.Error(err, "Failed to get Ingress object")
		return ctrl.Result{}, err
	}

	if instance.Spec.IngressClassName != nil && *instance.Spec.IngressClassName != NginxIngressClass {
		log.Info("Ingress class is not nginx. Ignoring")
		return ctrl.Result{}, nil
	}

	resources, errlist := p.converter.Convert(*instance)
	if len(errlist) > 0 {
		for _, err := range errlist {
			log.Error(err, "Failed to convert Ingress to Gateway resources")
		}
		return ctrl.Result{}, errlist.ToAggregate()
	}

	for _, v := range resources.Gateways {
		if err := controllerutil.SetControllerReference(instance, &v, p.Scheme); err != nil {
			return ctrl.Result{}, err
		}
		err = createOrUpdateGateway(ctx, &v, p.Client)
		if err != nil {
			log.Error(err, "Failed to create or update Gateway")
			return ctrl.Result{}, err
		}
	}
	for _, v := range resources.HTTPRoutes {
		if err := controllerutil.SetControllerReference(instance, &v, p.Scheme); err != nil {
			return ctrl.Result{}, err
		}
		err = createOrUpdateHttpRoute(ctx, &v, p.Client)
		if err != nil {
			log.Error(err, "Failed to create or update HTTPRoute")
			return ctrl.Result{}, err
		}
	}
	for _, v := range resources.GRPCRoutes {
		if err := controllerutil.SetControllerReference(instance, &v, p.Scheme); err != nil {
			return ctrl.Result{}, err
		}
		err = createOrUpdateGRPCRoute(ctx, &v, p.Client)
		if err != nil {
			log.Error(err, "Failed to create or update HTTPRoute")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func createOrUpdateHttpRoute(ctx context.Context, desired *gatewayv1.HTTPRoute, c client.Client) error {
	existing := desired.DeepCopy()
	_, err := controllerutil.CreateOrUpdate(ctx, c, existing, func() error {
		existing.Labels = desired.Labels
		existing.Annotations = common.MergeMaps(desired.Annotations, existing.Annotations)
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

func createOrUpdateGRPCRoute(ctx context.Context, desired *gatewayv1.GRPCRoute, c client.Client) error {
	existing := desired.DeepCopy()
	_, err := controllerutil.CreateOrUpdate(ctx, c, existing, func() error {
		existing.Labels = desired.Labels
		existing.Annotations = common.MergeMaps(desired.Annotations, existing.Annotations)
		existing.OwnerReferences = desired.OwnerReferences
		existing.Spec = desired.Spec
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to create or update GRPCRoute: %w", err)
	}
	existing.DeepCopyInto(desired)
	return nil
}

func createOrUpdateGateway(ctx context.Context, desired *gatewayv1.Gateway, c client.Client) error {
	existing := desired.DeepCopy()
	_, err := controllerutil.CreateOrUpdate(ctx, c, existing, func() error {
		existing.Labels = desired.Labels
		existing.Annotations = common.MergeMaps(desired.Annotations, existing.Annotations)
		existing.OwnerReferences = desired.OwnerReferences
		existing.Spec = desired.Spec
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to create or update Gateway: %w", err)
	}
	existing.DeepCopyInto(desired)
	return nil
}

func (p *Provider) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&networkingv1.Ingress{}).
		Owns(&gatewayv1.HTTPRoute{}).
		Owns(&gatewayv1.GRPCRoute{}).
		Complete(p)
}
