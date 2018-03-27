#!/bin/sh
export POSIXLY_CORRECT=1

glide install

root_path=$(echo "${0}" | sed -e 's|/\{1,\}[^/]*$||')
cd "${root_path}"

export CGO_ENABLED=0
go build .
#cp terraform-provider-helmcmd ~/.terraform.d/plugins/


