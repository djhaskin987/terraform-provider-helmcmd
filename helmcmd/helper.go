package helmcmd

/*
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
*/

import (
	"github.com/ghodss/yaml"
	"github.com/tdewolff/minify"
	minjson "github.com/tdewolff/minify/json"
	"regexp"
	"strings"
)

// This converts an input string which is valid YAML to minified JSON.  This is
// convenient because it makes the state smaller, which helps in case you are
// using the consul backend to store your data.  It also helps because it
// avoids subtle edge case bugs and formatting issues associated with slurping
// yaml in, then just spitting it out.  I have seen it cause problems when
// multi-line strings enter the picture.  Far better to normalize to json
func normalizeInput(input string) (string, error) {
	j, err := yaml.YAMLToJSON([]byte(input))
	if err != nil {
		return "", err
	}
	var b strings.Builder
	r := strings.NewReader(string(j))
	m := minify.New()
	m.AddFuncRegexp(regexp.MustCompile("[/+]json$"), minjson.Minify)
	if err := m.Minify("application/json", &b, r); err != nil {
		return "", err
	}
	return string(b.String()), nil
}

func AttemptNormalizeInput(input string) string {
	if normalized, err := normalizeInput(input); err != nil {
		return input
	} else {
		return normalized
	}
}
