# Copyright 2018 The Helm CMD TF Provider Authors, see the AUTHORS file.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

provider "helmcmd" {
    "alias" = "kubernetes-stable"
    "chart_source_type" = "repository"
    # When chart_source_type is `repository`, this is the name of the repo
    # whence to get the chart.
    "chart_source" = "stable"
    "debug" = true
    "kube_context" = "minikube"
}

resource "helmcmd_release" "plain" {
    provider = "helmcmd.kubernetes-stable"
    name = "plain-old-me"
    chart_name = "nginx-ingress"
    chart_version = "0.12.0"
    namespace = "kube-system"
}
