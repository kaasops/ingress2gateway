package ingressnginx

import (
	"fmt"

	"github.com/kubernetes-sigs/ingress2gateway/pkg/i2gw"
	"github.com/kubernetes-sigs/ingress2gateway/pkg/i2gw/providers/common"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

var (
	// RedirectHttpsFeature is the feature that redirects HTTP to HTTPS
	httpsRedirectScheme = "https"
	httpsRedirectPort   = gatewayv1.PortNumber(443)
)

func redirectHttpsFeature(ingresses []networkingv1.Ingress, gatewayResources *i2gw.GatewayResources) field.ErrorList {
	ruleGroups := common.GetRuleGroups(ingresses)
	for _, rg := range ruleGroups {
		for _, ir := range rg.Rules {
			ingress := ir.Ingress
			if redirectHttpsAnnotationEnabled(ingress) {
				namespaceedName := types.NamespacedName{Namespace: rg.Namespace, Name: httpsRedirectRouteName(rg.Name, rg.Host)}
				if _, ok := gatewayResources.HTTPRoutes[namespaceedName]; ok {
					continue
				}
				redirectRoute := httpsRedirectHTTPRoute(rg)
				redirectRoute.Spec.ParentRefs = []gatewayv1.ParentReference{}
				gatewayResources.HTTPRoutes[namespaceedName] = redirectRoute
			}
		}
	}
	return nil
}

func redirectHttpsAnnotationEnabled(ingress networkingv1.Ingress) bool {
	if c := ingress.Annotations["nginx.ingress.kubernetes.io/force-ssl-redirect"]; c == "true" {
		return true
	}
	return false
}

func httpsRedirectHTTPRoute(rg common.IngressRuleGroup) gatewayv1.HTTPRoute {
	httpRoute := gatewayv1.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{
			Name:      httpsRedirectRouteName(rg.Name, rg.Host),
			Namespace: rg.Namespace,
		},
		Spec: gatewayv1.HTTPRouteSpec{
			Hostnames: []gatewayv1.Hostname{gatewayv1.Hostname(rg.Host)},
			Rules: []gatewayv1.HTTPRouteRule{
				{
					Filters: []gatewayv1.HTTPRouteFilter{
						{
							Type: "RequestRedirect",
							RequestRedirect: &gatewayv1.HTTPRequestRedirectFilter{
								Scheme: &httpsRedirectScheme,
								Port:   &httpsRedirectPort,
							},
						},
					},
				},
			},
		},
	}
	httpRoute.SetGroupVersionKind(common.GatewayGVK)
	return httpRoute
}

func httpsRedirectRouteName(ingressName, host string) string {
	return fmt.Sprintf("%s-redirect-https", common.RouteName(ingressName, host))
}
