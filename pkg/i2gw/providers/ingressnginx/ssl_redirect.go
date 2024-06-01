package ingressnginx

import (
	"strings"

	"github.com/kubernetes-sigs/ingress2gateway/pkg/i2gw"
	"github.com/kubernetes-sigs/ingress2gateway/pkg/i2gw/providers/common"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/utils/ptr"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

var (
	httpsRedirectScheme = "https"
)

func sslRedirectFeature(ingresses []networkingv1.Ingress, gatewayResources *i2gw.GatewayResources) field.ErrorList {
	ruleGroups := common.GetRuleGroups(ingresses)
	for _, rg := range ruleGroups {
		for _, rule := range rg.Rules {
			if needsRedirectToHttps(rule.Ingress) {
				if rule.Ingress.Spec.Rules == nil {
					continue
				}
				key := types.NamespacedName{Namespace: rule.Ingress.Namespace, Name: common.RouteName(rg.Name, rg.Host)}
				httpRoute, ok := gatewayResources.HTTPRoutes[key]
				if !ok {
					continue
				}
				for i, rule := range httpRoute.Spec.Rules {
					rule.Filters = append(rule.Filters, gatewayv1.HTTPRouteFilter{
						Type: gatewayv1.HTTPRouteFilterRequestRedirect,
						RequestRedirect: &gatewayv1.HTTPRequestRedirectFilter{
							Scheme:     ptr.To(httpsRedirectScheme),
							StatusCode: ptr.To(int(301)),
						},
					})
					httpRoute.Spec.Rules[i] = rule
				}
			}
		}
	}
	return nil
}

func needsRedirectToHttps(ingress networkingv1.Ingress) bool {
	v, ok := ingress.Annotations["nginx.ingress.kubernetes.io/force-ssl-redirect"]
	if ok && strings.ToLower(v) == "true" {
		return true
	}
	if ingress.Spec.TLS != nil {
		v, ok = ingress.Annotations["nginx.ingress.kubernetes.io/ssl-redirect"]
		if ok && strings.ToLower(v) == "false" {
			return false
		}
		return true
	}
	return false
}
