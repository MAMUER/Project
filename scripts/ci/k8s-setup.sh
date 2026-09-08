#!/bin/bash
set -euo pipefail

setup_local_path_provisioner() {
	set -euo pipefail
	echo "Applying local-path-provisioner manifest..."
	kubectl apply -f configs/k8s/base/local-path-provisioner.yaml
	echo "Restarting local-path-provisioner to clear any stuck state..."
	kubectl rollout restart deployment/local-path-provisioner -n local-path-storage
	kubectl rollout status deployment/local-path-provisioner -n local-path-storage --timeout=60s
	echo "Waiting for local-path-provisioner pod to be ready..."
	kubectl -n local-path-storage rollout status deploy/local-path-provisioner --timeout=3m
	echo "Verifying StorageClass local-path exists and is ready..."
	for i in {1..30}; do
		if kubectl get storageclass local-path &>/dev/null; then
			echo "✅ StorageClass found"
			break
		fi
		echo "Attempt $i/30: waiting for StorageClass..."
		sleep 2
	done
	if ! kubectl get storageclass local-path -o jsonpath='{.metadata.annotations.storageclass\.kubernetes\.io/is-default-class}' | grep -q true; then
		echo "⚠️ StorageClass local-path is not default, patching..."
		kubectl patch storageclass local-path -p '{"metadata":{"annotations":{"storageclass.kubernetes.io/is-default-class":"true"}}}'
	fi
	echo "✅ local-path-provisioner and StorageClass are ready"
}

fix_local_path_rbac() {
	./scripts/ssh-retry.sh scp configs/k8s/scripts/fix-local-path.sh "${VPS_USER}@${VPS_HOST}:/tmp/fix-local-path.sh"
	./scripts/ssh-retry.sh scp configs/k8s/base/local-path-provisioner.yaml "${VPS_USER}@${VPS_HOST}:/tmp/local-path-provisioner.yaml"
	./scripts/ssh-retry.sh ssh "${VPS_USER}@${VPS_HOST}" "sudo chmod +x /tmp/fix-local-path.sh && sudo /tmp/fix-local-path.sh"
}

install_cert_manager() {
	echo "-> Installing cert-manager..."
	kubectl apply -f configs/k8s/base/cert-manager/
	echo "-> Waiting for cert-manager pods to be ready..."
	kubectl wait --namespace cert-manager --for=condition=ready pod --all --timeout=120s
	echo "✅ cert-manager installed successfully"

	echo "-> Creating Let's Encrypt ClusterIssuer..."
	kubectl apply -f configs/k8s/base/cert-manager/cluster-issuer.yaml
	echo "✅ ClusterIssuer created"

	echo "-> Waiting for ingress-nginx to be ready..."
	kubectl wait --namespace ingress-nginx --for=condition=ready pod --all --timeout=120s
	echo "✅ Ingress-nginx is ready"

	echo "-> Verifying ModSecurity configuration..."
	kubectl get configmap ingress-nginx-controller -n ingress-nginx -o jsonpath='{.data.modsecurity-snippet}' | grep -q "SecRuleEngine On" && echo "✅ ModSecurity enabled" || echo "❌ ModSecurity not enabled"

	echo "✅ Certificate and WAF configuration complete"
}

install_ingress_nginx() {
	echo "-> Installing ingress-nginx controller..."
	kubectl apply -f configs/k8s/base/ingress-nginx/

	echo "-> Waiting for ingress-nginx to be ready..."
	kubectl wait --namespace ingress-nginx \
		--for=condition=ready pod \
		--selector=app.kubernetes.io/name=ingress-nginx \
		--timeout=180s || {
		echo "❌ ingress-nginx not ready"
		kubectl -n ingress-nginx get pods,deploy,svc
		exit 1
	}

	echo "✅ ingress-nginx is ready"
	kubectl -n ingress-nginx get pods,deploy,svc
}

cleanup_traefik() {
	echo "-> Cleaning up Traefik from kube-system..."
	kubectl -n kube-system delete deployment traefik --ignore-not-found=true || true
	kubectl -n kube-system delete svc traefik --ignore-not-found=true || true
	kubectl -n kube-system get pods,svc,deploy | grep -i traefik || echo "Traefik already removed or not found"
	echo "✅ Traefik cleanup complete (non-fatal)"
}

main() {
	local cmd="${1:-all}"
	case "$cmd" in
	all)
		setup_local_path_provisioner
		fix_local_path_rbac
		install_cert_manager
		install_ingress_nginx
		cleanup_traefik
		;;
	setup_local_path_provisioner) setup_local_path_provisioner ;;
	fix_local_path_rbac) fix_local_path_rbac ;;
	install_cert_manager) install_cert_manager ;;
	install_ingress_nginx) install_ingress_nginx ;;
	cleanup_traefik) cleanup_traefik ;;
	*)
		echo "Unknown function: $cmd"
		exit 1
		;;
	esac
}

main "$@"
