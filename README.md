# Terraform Helm Provider

The main goal of this provider is to give the community a rock-solid, stable
way to use helm from terraform while easing the transition into terraform
for those who have historically just used helm.

This [terraform](https://www.terraform.io/) [helm](https://helm.sh/) provider
is different from the main [helm
provider](https://github.com/mcuadros/terraform-provider-helm) in two main
ways.

First, it relies on the helm CLI under the covers. This implies that you will
need to configure kubectl and install kubectl and helm on the same
machine that terraform runs. It *also* implies that when error messages happen,
they will be familiar to those who have already used helm in the past. This
eases onboarding into Terraform. This is not better or worse than using the
helm golang libraries to talk to helm; it is simply different, and hopefully
useful for some :) It also allows the provider to call `helm upgrade --install`
for *both* creation and update of the resource.

Second, it strictly adheres to the principle that everything in the resource
`helm_release` is associated with a chart
[lifecycle](https://www.terraform.io/docs/internals/lifecycle.html). This will
hopefully make this helm provider more stable and usable. As a side effect,
this provider supports `terraform import`, again to ease onboarding from
previous terraform-less processes.

In the initial release, the `helm upgrade` and `helm delete` arguments used are
hard-coded.  This may change in the future, but this project is not opposed to
baking opinions into how helm should be called from terraform.

# Install

Go grab the executable for your platform from this project's
[releases](https://github.com/djhaskin987/terraform-provider-helmcmd/releases)
and put it in the `~/.terraform.d/plugins` folder.

# Contribute

PLEASE! This provider is in early stages, but hopes to be useful immediately.
If there are any issues, please do not hesitate to start discussions via github
issues and contribute via PRs.

# Use

Please see the `examples/` folder for examples on how to use this provider.


## Provider Configuration

```hcl
provider "helmcmd_release" {
  kube_context = "..."
  chart_source_type: "..."
  chart_source = "..."
  debug =  true/false
}
```
Each provider is associated with a chart source. This was necessary to preserve
the "lifecycle-ness" of the chart name in the helm release. The name of the
helm chart can be read from helm using `helm list`, but its source cannot,
so that must live in the provider.

If the `chart_source_type` is "filesystem", `chart_source` must be a directory
which has charts in it, with each chart being in a directory named the same
as the name of the chart. When used, filesystem charts will have `helm dep
update` run on them, so that they don't have to be fully built first. This
allows operators to create terraform modules strictly in git and depend
on other TF helm modules using the git repository link, no helm repository
required.

If the `chart_source_type` is "repository", `chart_source` must be
the name of one of the helm repositories previously configured in the helm CLI.
Example: `stable`.

The other provider parameters coincide with global helm CLI options: `debug`
for `--debug` and `kube_context` for `--kube-context`, respectively.

## Resource Configuration

This Provider currently only presents a single resource, `helmcmd_release`.
It has the following parameters:

```hcl
resource "helmcmd_release" "myrelease" {
  name = "helm-release-name"
  chart_name = "chart-name"
  chart_verion = "0.2.0"
  namespace = "default"
  overrides = "{}"
}
```

The parameters `namespace` and `overrides` above are optional and
are shown above with their default values. Each of the above
parameters are can be read from the output of `helm list`, except `overrides`
which can be read from `helm get values`. Indeed, this is how they are
read into terraform when `terraform refresh` is called.

The `name` parameter is the name of the helm release. If it is changed, a
new resource is forced, so that the old release will be deleted. At the time
of writing, this resource deletes releases using `helm delete --purge`.

The `chart_name` and `chart_version` parameters are both required.
`chart_name` must be the same as the name of the chart as it will appear in a
release entry in the output of the `helm list` command.

The `overrides` parameter is a YAML string used as input to the `-f` parameter
when `helm upgrade` is called. It is normalized to minified JSON in the
internal state to ensure that two equivalent YAML documents are compared
correctly regardless of whitespace differences, and also to avoid errors
associated with slurping YAML in and spitting it back out as YAML. (The
original author has had run-ins with this when using Helm's on `toYaml` in
combination with multi-line strings.)

## Terraform Import

Internally, terraform uses the name of the release as the IDs of the resources.
Existing helm releases can be imported by feeding the name of the release to
`terraform import`.

For example, for the following resource:

```hcl
provider "helmcmd" {
    "alias" = "plain"
    "chart_source_type" = "filesystem"
    # When chart_source_type is the filesystem, a directory is specified in
    # chart_source.  The name of the directory housing a chart (e.g., "plain")
    # must *match* the name of the chart that it houses.
    "chart_source" = "${path.module}/charts"
    "debug" = true
    "kube_context" = "minikube"
}

resource "helmcmd_release" "plain" {
    provider = "helmcmd.plain"
    name = "plain-old-me2"
    chart_name = "plain"
    chart_version = "0.1.0"
    overrides = "${file("${path.module}/overrides/default.yaml")}"
}
```

You could import an existing release like this:

```bash
terraform import helmcmd_release.plain plain-old-me2
```


It should also be noted that, intentionally, `helm upgrade --install` is used
for both creation and update of the resource internally. This means that
terraform import is not strictly needed, but is possible.

# Copyright Information

Copyright 2018 The Helm CMD TF Provider Authors, see the AUTHORS file.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

