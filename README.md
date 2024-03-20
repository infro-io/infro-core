<p align="center">
<img src="https://raw.githubusercontent.com/infro-io/infro-core/main/.github/images/banner.png?sanitize=true" alt="infro" width="100%">
</p>

<p align=center>
<a href="https://hub.docker.com/r/infrolabs/infro-core">
  <img alt="GitHub Action Status" src="https://img.shields.io/docker/v/infrolabs/infro-core">
</a>
<a href="https://github.com/infro-io/infro-core/actions/workflows/main.yml">
  <img alt="GitHub Action Status" src="https://github.com/infro-io/infro-core/actions/workflows/main.yml/badge.svg">
</a>
</p>


## What is it

Infro is a GitOps tool which helps you understand how your changes will affect your infrastructure, before you merge.
By integrating with different IaC providers, Infro provides a clear, holistic view of your changes on pull requests:

<p align="center">
<img src="https://raw.githubusercontent.com/infro-io/infro-core/main/.github/images/diff.png?sanitize=true" alt="diff" width="60%">
</p>

## Features

Infro supports running dry runs on different IaC providers and publishing results to different version control systems:

|                 | GitHub Action | Self-hosted | Infro Cloud |
|-----------------|---------------|-------------|-------------|
| Argo CD diffs   | ✅             | ✅           | ✅           |
| Terraform diffs | ✅             | 🚧          | 🚧          |
| AWS CDK diffs   | 🚧            | 🚧          | 🚧          |
| GitHub comments | ✅             | ✅           | ✅           |
| GitLab comments | 🚧            | 🚧          | 🚧          |
| No-code         | ❌             | ❌           | ✅           |
| PR Check status | ✅             | ❌           | ✅           |

## Why Infro?

Nearly all IaC tools provide an option to perform diffs and dry runs,
but they usually rely on you manually running them.
Many companies have created custom solutions to publish PR comments for their IaC of choice,
but there hasn't yet been a generic solution.

<center>
<table>
  <tr>
      <th><h3>🔒</h3><h3>Secure</h3></th>
      <th><h3>🧰</h3><h3>Extensible</h3></th>
    </tr>
    <tr>
      <td width="33%"><sub>Infro doesn't require you to expose your infrastructure to the public internet or give third-party permissions.</sub></td>
      <td width="50%"><sub>One solution should be able to double check your changes across all your infra. Is your IaC or VCS not supported? It should be easy to integrate.</sub></td>
    </tr>
  <tr>
    <th><h3>🤖️</h3><h3>Automated</h3></th>
    <th><h3>🎧</h3><h3>Reduce Noise</h3></th>
  </tr>
  <tr>
    <td width="50%"><sub>Infro doesn't rely on manually intervention to perform dry runs across different infrastructure. Just open a pull request, and let Infro take care of the rest.</sub></td>
    <td width="33%"><sub>You don't always need to see every changed line, just the ones that matter. Infro is on a mission to point out the changes that matter.  </sub></td>
  </tr>
</table>
</center>

## How it works

<p align="center">
<img src="https://raw.githubusercontent.com/infro-io/infro-core/main/.github/images/integrations.svg?sanitize=true" alt="infro" width="100%">
</p>

Infro integrates with different version control systems (VCSs) and IaC providers (i.e. deployers).
Infro detects when a new commit is added to a VCS pull request.
Depending on the installation, this happens either by [polling for changes](#poll-mode) or by an [event trigger](#event-driven-mode).
Infro determines how the commit will change the live infrastructure by running a dry run against all of the IaC providers (i.e. deployers) defined in the [configuration yaml](#configuration),
and comment the diffs back to the pull request.

#### Poll Mode
Self-hosted deployments run in poll mode.
Infro will continuously poll the VCS API (e.g. Github) to scan for new pull requests commits under the configured organization or user (i.e. owner).
Therefore, you don't need to expose your infrastructure to a third-party.

#### Event-Driven Mode
Some deployments run in event-driven mode such as GitHub Actions and Infro Cloud.
When a new commit is added to a pull request, Infro is triggered to perform dry runs for that revision.

## Install

### As a Github Action
Installing as a Github Action to have Infro comment diffs on each pull request in that repository.
Add a job to your GitHub action yaml that is triggered when a pull_request is opened.
Here's an example configuration for running diffs against an Argo CD cluster:

```yaml
on:
  pull_request:
    branches:
      - main

jobs:
  comment-diffs:
    name: comment diffs
    runs-on: ubuntu-latest
    permissions:
      pull-requests: write
    steps:
      - uses: infro-io/comment-diffs-action@v1.4
        with:
          repo: ${{ github.repository }}
          revision: ${{ github.sha }}
          pull-number: ${{ github.event.number }}
          config: |
            deployers:
              - type: argocd
                name: my-argo-cluster
                endpoint: ${{ secrets.ARGOCD_ENDPOINT }}
                authtoken: ${{ secrets.ARGOCD_TOKEN }}
            vcs:
              type: github
              authtoken: ${{ secrets.GITHUB_TOKEN }}
```

You can use the [example-helm repo](https://github.com/infro-io/example-helm/pull/4) for reference.
The action requires `pull-requests: write` permissions to comment.
You must also pass the `repo`, `revision`, and `pull-number` from github context.
`GITHUB_TOKEN` is provided automatically, but you must provide additional secrets, such as `ARGOCD_TOKEN`, in repository `secrets`.
More information about the configuration [here](#configuration).

### Self-hosted

Infro can be installed into your Kubernetes cluster by using kubectl:
```shell
kubectl create namespace infro
kubectl apply -n infro -f https://raw.githubusercontent.com/infro-io/infro-core/main/deploy/install.yaml
```
or you can install it with Kustomize:
```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namespace: infro
resources:
- https://raw.githubusercontent.com/infro-io/infro-core/main/deploy/install.yaml
```

You can also try this out locally by using [Tilt](https://tilt.dev/).
Check out this repository and open the `Tiltfile`.
Replace the `ARGOCD_ENDPOINT`, `ARGOCD_TOKEN`, `GITHUB_TOKEN`, and `GITHUB_OWNER` values, then run:

```shell
tilt up
```
Then open a pull request for a repository managed by your Argo CD and watch the magic happen.
More information about the configuration [here](#configuration).

### Infro Cloud

There is no code or deployment required if you choose to use Infro Cloud.
All setup is performed in the UI at [infro.io](https://infro.io).
1. Configure your organization https://infro.io/docs/configure-org
2. Registering your Argo CD instance https://infro.io/docs/register-argo-cd

## Configuration

- The config yaml structure:
```yaml
deployers:
  - type: argocd
    name: <ARBITRARY_NAME>
    authtoken: <ARGOCD_TOKEN>
    endpoint: <ARGOCD_ENDPOINT>
  - type: terraform
    workdir: <TERRAFORM_WORKDIR>
vcs:
  - type: github
    authtoken: <GITHUB_TOKEN>
```

### `argocd`:
- `endpoint`: The Argo CD server address without the protocol (e.g. `http`).
For example, the default endpoint in Kubernetes clusters should be `argocd-server.argocd.svc.cluster.local`. 
- `authtoken`: an [Argo CD automation token](https://argo-cd.readthedocs.io/en/stable/operator-manual/security/#authentication). See how to generate a secure token:
<details id="argocd-token-generation">
<summary>How to generate an Argo CD token</summary>

###### 1. Add an account to your Argo CD ConfigMap.

The account needs the [`apiKey`](https://argo-cd.readthedocs.io/en/stable/operator-manual/user-management/) capability in order to generate an auth token.
If you already have an account with this capability, you can skip this step.
The account can be added by modifying [`argocd-cm`](https://argo-cd.readthedocs.io/en/stable/operator-manual/argocd-cm-yaml/) ConfigMap:

~~~yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: argocd-cm
  namespace: argocd
  labels:
    app.kubernetes.io/name: argocd-cm
    app.kubernetes.io/part-of: argocd
data:
  accounts.infro: "apiKey"
~~~

###### 2. Add an account policy.

The account also needs permissions to read `applications` and `projects` to perform `app diff`.
The account policy can be added by modifying [`argocd-rbac-cm`](https://argo-cd.readthedocs.io/en/stable/operator-manual/argocd-cm-yaml/) ConfigMap:

~~~yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: argocd-rbac-cm
  namespace: argocd
  labels:
    app.kubernetes.io/name: argocd-rbac-cm
    app.kubernetes.io/part-of: argocd
data:
  policy.csv: |
    p, role:readonly, applications, get, *, allow
    p, role:readonly, projects, get, *, allow
    g, infro, role:readonly
~~~

###### 3. Generate an auth token for the account.

Visit your ArgoCD UI and go to *Settings > Accounts*. Under Tokens, click *Generate New*.

</details>

### `terraform`:
- `workdir`: The [Terraform working directory](https://developer.hashicorp.com/terraform/cli/init).

### `github`:
- `authtoken`: a GitHub personal access token. See how to create one [here](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens#creating-a-fine-grained-personal-access-token).

# Contributing

All contributions are welcome!

### Development

Commands:
```shell
$ make help
build-docker: Build Docker image
build-go: Build go binary
format: Format code based on linter configuration
help: Show help for each of the Makefile recipes.
lint: Run linters
test-integration: Run integration tests
test-unit: Run unit tests
```
