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

func useRegexFeature(ingresses []networkingv1.Ingress, gatewayResources *i2gw.GatewayResources, services map[types.NamespacedName]*corev1.Service) field.ErrorList {
	ruleGroups := common.GetRuleGroups(ingresses)
	for _, rg := range ruleGroups {
		for _, rule := range rg.Rules {
			if useRegexAnnotation(rule.Ingress) {
				if rule.Ingress.Spec.Rules == nil {
					continue
				}
				key := types.NamespacedName{Namespace: rule.Ingress.Namespace, Name: rg.Name}
				httpRoute, ok := gatewayResources.HTTPRoutes[key]
				if !ok {
					continue
				}
				for _, rule := range httpRoute.Spec.Rules {
					for _, path := range rule.Matches {
						path.Path.Type = common.PtrTo(gatewayv1.PathMatchRegularExpression)
					}
				}
			}
		}
	}
	return nil
}

func useRegexAnnotation(ingress networkingv1.Ingress) bool {
	v, ok := ingress.Annotations["nginx.ingress.kubernetes.io/use-regex"]
	if ok && strings.ToLower(v) == "true" {
		return true
	}
	return false
}
