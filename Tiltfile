
# extensions
load('ext://secret', 'secret_create_generic', 'secret_from_dict', 'secret_yaml_registry')

# deploy local infro
config = '''
deployers:
  - type: argocd
    name: my-beta-cluster
    endpoint: {argocdEndpoint}
    authtoken: {argocdToken}
vcs:
  type: github
  authtoken: {githubToken}
'''.format(
  argocdEndpoint=os.environ['ARGOCD_ENDPOINT'],
  argocdToken=os.environ['ARGOCD_TOKEN'],
  githubToken=os.environ['GITHUB_TOKEN'],
)
k8s_yaml(secret_from_dict("infro-secrets", "infro", inputs={
    'config': config,
    'owner': os.environ['GITHUB_OWNER'], # the github org or user
}))
k8s_yaml('./deploy/install.yaml')