
# deploy local argocd
k8s_yaml(kustomize('./test/argocd'))

# expose argocd api server
k8s_resource(workload='argocd-server', port_forwards=8080)