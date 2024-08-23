URL=https://kaasops.github.io/ingress2gateway/helm

# Run linter for helm chart
lint:
	helm lint charts/ingress2gateway

# Package chart into zip file
package:
	helm package charts/* -d packages

# Create index file (use only for initial setup)
index:
	helm repo index --url ${URL} .