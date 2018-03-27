provider "helmcmd" {
    "alias" = "plain"
    "chart_source_type" = "filesystem"
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
