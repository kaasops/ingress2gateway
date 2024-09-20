package ingressnginx

import (
	"strings"

	"github.com/kubernetes-sigs/ingress2gateway/pkg/i2gw"
	"github.com/kubernetes-sigs/ingress2gateway/pkg/i2gw/providers/common"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func grpcBackendProtocolFeature(ingresses []networkingv1.Ingress, gatewayResources *i2gw.GatewayResources, services map[types.NamespacedName]*corev1.Service) field.ErrorList {
	var errs field.ErrorList
	ruleGroups := common.GetRuleGroups(ingresses)
	gatewayResources.GRPCRoutes = make(map[types.NamespacedName]gatewayv1.GRPCRoute)
	for _, rg := range ruleGroups {
		if len(rg.Rules) == 0 {
			continue
		}
		v, ok := rg.Rules[0].Ingress.Annotations["nginx.ingress.kubernetes.io/backend-protocol"]
		if ok && strings.ToLower(v) == "grpc" {
			key := types.NamespacedName{Namespace: rg.Namespace, Name: rg.Name}
			httpRoute, ok := gatewayResources.HTTPRoutes[key]
			if !ok {
				errs = append(errs, field.NotFound(field.NewPath("HTTPRoute"), key))
			}
			grpcRoute := toGRPCRoute(httpRoute)
			grpcRoute.SetGroupVersionKind(common.GRPCRouteGVK)
			delete(gatewayResources.HTTPRoutes, key)
			gatewayResources.GRPCRoutes[key] = grpcRoute
		}
	}
	return errs

}

func toGRPCRoute(httpRoute gatewayv1.HTTPRoute) gatewayv1.GRPCRoute {
	rules := make([]gatewayv1.GRPCRouteRule, 0, len(httpRoute.Spec.Rules))
	for _, rule := range httpRoute.Spec.Rules {
		rules = append(rules, toGRPCRule(rule))
	}
	return gatewayv1.GRPCRoute{
		ObjectMeta: httpRoute.ObjectMeta,
		Spec: gatewayv1.GRPCRouteSpec{
			Hostnames: httpRoute.Spec.Hostnames,
			CommonRouteSpec: gatewayv1.CommonRouteSpec{
				ParentRefs: httpRoute.Spec.ParentRefs,
			},
			Rules: rules,
		},
	}
}

func toGRPCRule(httpRule gatewayv1.HTTPRouteRule) gatewayv1.GRPCRouteRule {
	grpcBackendRefs := make([]gatewayv1.GRPCBackendRef, 0, len(httpRule.BackendRefs))
	for _, ref := range httpRule.BackendRefs {
		grpcBackendRefs = append(grpcBackendRefs, toGRPCBackendRef(ref))
	}

	return gatewayv1.GRPCRouteRule{
		BackendRefs: grpcBackendRefs,
	}
}

func toGRPCBackendRef(httpBackendRefs gatewayv1.HTTPBackendRef) gatewayv1.GRPCBackendRef {
	return gatewayv1.GRPCBackendRef{
		BackendRef: httpBackendRefs.BackendRef,
	}
}
