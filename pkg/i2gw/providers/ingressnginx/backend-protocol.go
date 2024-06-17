package ingressnginx

import (
	"strings"

	"github.com/kubernetes-sigs/ingress2gateway/pkg/i2gw"
	"github.com/kubernetes-sigs/ingress2gateway/pkg/i2gw/providers/common"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	gatewayv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

func backendProtocolFeature(ingresses []networkingv1.Ingress, gatewayResources *i2gw.GatewayResources) field.ErrorList {
	var errs field.ErrorList
	ruleGroups := common.GetRuleGroups(ingresses)
	gatewayResources.GRPCRoutes = make(map[types.NamespacedName]gatewayv1alpha2.GRPCRoute)
	for _, rg := range ruleGroups {
		if len(rg.Rules) == 0 {
			continue
		}
		v, ok := rg.Rules[0].Ingress.Annotations["nginx.ingress.kubernetes.io/backend-protocol"]
		if ok && strings.ToLower(v) == "grpc" {
			key := types.NamespacedName{Namespace: rg.Namespace, Name: common.RouteName(rg.Name, rg.Host)}
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

func toGRPCRoute(httpRoute gatewayv1.HTTPRoute) gatewayv1alpha2.GRPCRoute {
	return gatewayv1alpha2.GRPCRoute{
		ObjectMeta: httpRoute.ObjectMeta,
		Spec: gatewayv1alpha2.GRPCRouteSpec{
			Hostnames: httpRoute.Spec.Hostnames,
			CommonRouteSpec: gatewayv1alpha2.CommonRouteSpec{
				ParentRefs: httpRoute.Spec.ParentRefs,
			},
		},
	}
}
