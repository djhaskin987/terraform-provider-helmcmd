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
	"bytes"
	"fmt"
	"github.com/dogenzaka/tsv"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

/* TODO:
   - Incorporate helm dep update
*/
type HelmCmd struct {
	Debug                   bool
	Home                    string
	Host                    string
	KubeContext             string
	Kubeconfig              string
	TillerConnectionTimeout int
	TillerNamespace         string
	Timeout                 int
	ChartSourceType         string
	ChartSource             string
}

type HelmRelease struct {
	Name         string
	ChartName    string
	ChartVersion string
	Namespace    string
	Overrides    string
}

var ErrHelmNotExist error = fmt.Errorf("Couldn't find release")
var ErrUnsuccessfulDeploy error = fmt.Errorf("Unsuccessful deploy")

func run(cmd *exec.Cmd) error {
	stderr := &bytes.Buffer{}
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		cmdStr := cmd.Path + " " + strings.Join(cmd.Args, " ")
		if stderr.Len() == 0 {
			return fmt.Errorf("%s: %v", cmdStr, err)
		}
		return fmt.Errorf("%s %v: %s", cmdStr, err, stderr.Bytes())
	}
	return nil
}

// Return global arguments that do not affect output format.
// This is used in e.g. helmReadFromList, which needs to get `helm list`
// output but also needs to put other behavioral global arguments
// into helm as well. So, helmReadFromList *needs* to call
// `helm list` without the `debug` flag even if it was
// specified.
func (c *HelmCmd) behavioralGlobalArgs() []string {
	args := []string{}

	if c.Home != "" {
		args = append(args, "--home", c.Home)
	}
	if c.Host != "" {
		args = append(args, "--host", c.Host)
	}
	if c.KubeContext != "" {
		args = append(args, "--kube-context", c.KubeContext)
	}
	if c.Kubeconfig != "" {
		args = append(args, "--kubeconfig", c.Kubeconfig)
	}
	if c.TillerConnectionTimeout >= 0 {
		args = append(args, "--tiller-connection-timeout", strconv.Itoa(c.TillerConnectionTimeout))
	}
	if c.TillerNamespace != "" {
		args = append(args, "--tiller-namespace", c.TillerNamespace)
	}
	return args
}

func (c *HelmCmd) globalArgs() []string {
	args := c.behavioralGlobalArgs()
	if c.Debug {
		args = append(args, "--debug")
	}
	return args
}

func (r *HelmRelease) Validate() error {
	if r.ChartName == "" {
		return fmt.Errorf("Chart name is unset: %v", r)
	}
	if r.ChartVersion == "" {
		return fmt.Errorf("Chart version is unset: %v", r)
	}
	if r.Namespace == "" {
		return fmt.Errorf("Namespace is unset: %v", r)
	}
	return nil
}

func (c *HelmCmd) Validate() error {
	if c.ChartSourceType == "filesystem" {
		if st, err := os.Stat(c.ChartSource); os.IsNotExist(err) {
			return fmt.Errorf("Chart source must be an existent directory")

		} else if !st.IsDir() {
			return fmt.Errorf("Chart source must be an existent directory")
		}
	} else if c.ChartSourceType != "repository" {
		return fmt.Errorf("Chart source type not specified correctly, it must either be `repository` or `filesystem`")
	}
	return nil
}

type HelmReleaseInfo struct {
	Name         string
	Revision     int
	LastUpdated  time.Time
	Status       string
	ChartName    string
	ChartVersion string
	Namespace    string
}

func (c *HelmCmd) deleteRaw(release *HelmRelease) error {
	deleteArgs := c.globalArgs()
	deleteArgs = append(deleteArgs, "delete")
	if c.Timeout >= 0 {
		deleteArgs = append(deleteArgs, "--timeout", strconv.Itoa(c.Timeout))
	}
	deleteArgs = append(deleteArgs, "--purge", release.Name)
	stdout := &bytes.Buffer{}
	cmd := exec.Command("helm", deleteArgs...)
	cmd.Stdout = stdout
	if err := run(cmd); err != nil {
		return err
	}
	log.Printf("%s\n", stdout.String())
	return nil
}

func (c *HelmCmd) Upgrade(release *HelmRelease) error {
	if err := release.Validate(); err != nil {
		return err
	}

	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	upgradeArgs := c.globalArgs()
	upgradeArgs = append(upgradeArgs, "upgrade", "--install", "--devel",
		"--wait", "-f", "-")
	if c.Timeout >= 0 {
		upgradeArgs = append(upgradeArgs, "--timeout", strconv.Itoa(c.Timeout))
	}
	upgradeArgs = append(upgradeArgs, "--version", release.ChartVersion)
	upgradeArgs = append(upgradeArgs, "--namespace", release.Namespace)
	upgradeArgs = append(upgradeArgs, release.Name)

	if c.ChartSourceType == "repository" {
		upgradeArgs = append(upgradeArgs, "--repo", c.ChartSource)
		upgradeArgs = append(upgradeArgs, release.ChartName)
	} else if c.ChartSourceType == "filesystem" {
		chartLocation := filepath.Join(c.ChartSource, release.ChartName)
		repoUpdateArgs := c.globalArgs()
		repoUpdateArgs = append(repoUpdateArgs, "repo", "update")
		helmRepoUpdate := exec.Command("helm", repoUpdateArgs...)
		if err := run(helmRepoUpdate); err != nil {
			return err
		}
		helmDepUpdateArgs := c.globalArgs()
		helmDepUpdateArgs = append(helmDepUpdateArgs, "dependency", "update",
			chartLocation)
		helmDepUpdate := exec.Command("helm", helmDepUpdateArgs...)
		if err := run(helmDepUpdate); err != nil {
			return err
		}
		upgradeArgs = append(upgradeArgs, chartLocation)
	}

	cmd := exec.Command("helm", upgradeArgs...)
	cmd.Stderr = stderr
	cmd.Stdout = stdout
	cmd.Stdin = strings.NewReader(release.Overrides)

	commandDisplay := fmt.Sprintf("helm %s", strings.Join(upgradeArgs, " "))
	if err := cmd.Run(); err != nil {
		if stderr.Len() == 0 {
			return fmt.Errorf("%s: %v", commandDisplay, err)
		}
		return fmt.Errorf("%s %v: %s", commandDisplay, err, stderr.Bytes())
	}
	log.Printf("%s\n", stdout.String())

	results, err := c.helmReadFromList(release)
	if err != nil {
		return err
	}
	if results.Status != "DEPLOYED" {
		return ErrUnsuccessfulDeploy
	}

	return nil
}

func (c *HelmCmd) Delete(release *HelmRelease) error {
	if err := release.Validate(); err != nil {
		return err
	}
	return c.deleteRaw(release)
}

func (c *HelmCmd) Read(release *HelmRelease) error {

	results, err := c.helmReadFromList(release)
	if err != nil {
		return err
	}
	if results.Status == "DELETED" {
		return ErrHelmNotExist
	}
	if results.Status != "DEPLOYED" {
		return ErrUnsuccessfulDeploy
	}

	release.Name = results.Name
	release.ChartName = results.ChartName
	release.ChartVersion = results.ChartVersion
	release.Namespace = results.Namespace

	stdout := &bytes.Buffer{}
	cmdArgs := c.behavioralGlobalArgs()
	cmdArgs = append(cmdArgs, "get", "values", release.Name)
	cmd := exec.Command("helm", cmdArgs...)
	cmd.Stdout = stdout
	if err := run(cmd); err != nil {
		return fmt.Errorf("Couldn't read overrides for release %s: %s",
			release.Name, err.Error())
	}
	if overrides, err := normalizeInput(stdout.String()); err != nil {
		return err
	} else {
		release.Overrides = overrides
	}

	return nil
}

type HelmReleaseInfoRow struct {
	Name        string `tsv:"NAME"`
	Revision    string `tsv:"REVISION"`
	LastUpdated string `tsv:"UPDATED"`
	Status      string `tsv:"STATUS"`
	Chart       string `tsv:"CHART"`
	Namespace   string `tsv:"NAMESPACE"`
}

func helmReadRow(release *HelmRelease, currentRow *HelmReleaseInfoRow) (*HelmReleaseInfo, error) {
	results := &HelmReleaseInfo{}
	results.Name = currentRow.Name
	if revision, err := strconv.Atoi(currentRow.Revision); err != nil {
		return nil, fmt.Errorf("Couldn't read revision for release %s: %s",
			err.Error())
	} else {
		results.Revision = revision
	}

	zname, zoffset := time.Now().Zone()
	zloc := time.FixedZone(zname, zoffset)
	if t, err := time.ParseInLocation(
		"Mon Jan  2 15:04:05 2006", currentRow.LastUpdated, zloc); err != nil {
		return nil, fmt.Errorf("Couldn't read updated time for release %s: %s",
			err.Error())
	} else {
		results.LastUpdated = t
	}

	results.Status = currentRow.Status
	ChartRegex := regexp.MustCompile("^([a-z]([-a-z0-9]*[a-z0-9])?)-([0-9]+\\.[0-9]+\\.[0-9]+.*)$")
	chartStr := currentRow.Chart
	matches := ChartRegex.FindStringSubmatch(chartStr)
	if len(matches) < 4 {
		return nil, fmt.Errorf("Couldn't parse chart name from version in release `%s`: `%s`",
			release.Name, chartStr)
	} else {
		results.ChartName = matches[1]
		results.ChartVersion = matches[3]
	}
	results.Namespace = currentRow.Namespace
	return results, nil
}

func (c *HelmCmd) helmReadFromList(release *HelmRelease) (*HelmReleaseInfo, error) {
	stdout := &bytes.Buffer{}
	cmdArgs := c.behavioralGlobalArgs()
	cmdArgs = append(cmdArgs, "list", "-a")
	cmd := exec.Command("helm", cmdArgs...)
	cmd.Stdout = stdout
	if err := run(cmd); err != nil {
		return nil, err
	}
	leftSpaces := regexp.MustCompile("(?m)^[ ]+")
	trimmedOutput := leftSpaces.ReplaceAllString(stdout.String(), "")

	middleSpaces := regexp.MustCompile("[ ]*\t+[ ]*")
	slimmedOutput := middleSpaces.ReplaceAllString(trimmedOutput, "\t")

	rightSpaces := regexp.MustCompile("(?m)[ ]+$")
	cleanOutput := rightSpaces.ReplaceAllString(slimmedOutput, "")

	log.Printf("Output:\n%s\n", cleanOutput)
	if cleanOutput == "" {
		return nil, ErrHelmNotExist
	}
	currentRow := HelmReleaseInfoRow{}
	r := strings.NewReader(cleanOutput)
	parser, _ := tsv.NewParser(r, &currentRow)

	for {
		eof, err := parser.Next()
		if eof {
			return nil, ErrHelmNotExist
		}
		if err != nil {
			return nil, err
		}
		log.Println("Current row: ", currentRow)
		if currentRow.Name == release.Name {
			if result, err := helmReadRow(release, &currentRow); err != nil {
				return nil, err
			} else {
				return result, nil
			}
		}
	}
}
