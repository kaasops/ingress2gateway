package ingressnginx

import (
	"github.com/kubernetes-sigs/ingress2gateway/pkg/i2gw/providers/common"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func implementationSpecificHTTPPathTypeMatch(path *gatewayv1.HTTPPathMatch) {
	path.Type = common.PtrTo(gatewayv1.PathMatchPathPrefix)
}
