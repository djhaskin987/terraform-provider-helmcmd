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
