package main // import "github.com/djhaskin987/terraform-provider-helmcmd"

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
	"fmt"
	"github.com/djhaskin987/terraform-provider-helmcmd/helmcmd"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/plugin"
	"github.com/hashicorp/terraform/terraform"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: func() terraform.ResourceProvider {
			return &schema.Provider{
				Schema: map[string]*schema.Schema{
					"kube_context": {
						Type:        schema.TypeString,
						Optional:    true,
						Default:     "",
						Description: "Kube context to use in config",
					},
					"chart_source_type": {
						Type:     schema.TypeString,
						Optional: true,
						Default:  "repository",
						Description: "Must be set either to `repository` " +
							"or `filesystem`.",
					},
					"chart_source": {
						Type:     schema.TypeString,
						Required: true,
						Description: "If `chart_source_type` is set to " +
							"`filesystem`, this is a path where charts can " +
							"be found by name. If it is set to `repository`," +
							" this is the specifier for a chart URL.",
					},
					"debug": {
						Type:        schema.TypeBool,
						Optional:    true,
						Default:     false,
						Description: "Whether to turn on helm debug output",
					},
				},
				ResourcesMap: map[string]*schema.Resource{
					"helmcmd_release": resourceManifest(),
				},
				ConfigureFunc: providerConfigure,
			}
		},
	})
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	result := &helmcmd.HelmCmd{
		KubeContext:     d.Get("kube_context").(string),
		ChartSourceType: d.Get("chart_source_type").(string),
		ChartSource:     d.Get("chart_source").(string),
		Debug:           d.Get("debug").(bool),
	}

	if result.ChartSourceType != "repository" &&
		result.ChartSourceType != "filesystem" {
		return result, fmt.Errorf(
			"Chart source type must be `repository` or `filesystem`, "+
				"got this instead: %s\n",
			result.ChartSourceType)
	}

	if err := result.Validate(); err != nil {
		return result, err
	}
	return result, nil
}

func resourceManifest() *schema.Resource {
	return &schema.Resource{
		Create: resourceManifestCreate,
		Read:   resourceManifestRead,
		Update: resourceManifestUpdate,
		Delete: resourceManifestDelete,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"chart_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"chart_version": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"namespace": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "default",
			},
			"overrides": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "{}",
				StateFunc: func(thing interface{}) string {
					return helmcmd.AttemptNormalizeInput(thing.(string))
				},
			},
		},
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
	}
}

func resourceManifestCreate(d *schema.ResourceData, m interface{}) error {
	cmd := m.(*helmcmd.HelmCmd)
	d.SetId(d.Get("name").(string))
	release := &helmcmd.HelmRelease{
		Name:         d.Get("name").(string),
		ChartName:    d.Get("chart_name").(string),
		ChartVersion: d.Get("chart_version").(string),
		Namespace:    d.Get("namespace").(string),
		Overrides:    d.Get("overrides").(string),
	}

	return cmd.Upgrade(release)
}

func resourceManifestUpdate(d *schema.ResourceData, m interface{}) error {
	cmd := m.(*helmcmd.HelmCmd)

	release := &helmcmd.HelmRelease{
		Name:         d.Get("name").(string),
		ChartName:    d.Get("chart_name").(string),
		ChartVersion: d.Get("chart_version").(string),
		Namespace:    d.Get("namespace").(string),
		Overrides:    d.Get("overrides").(string),
	}

	return cmd.Upgrade(release)
}

func resourceManifestDelete(d *schema.ResourceData, m interface{}) error {
	cmd := m.(*helmcmd.HelmCmd)

	release := &helmcmd.HelmRelease{
		Name:         d.Get("name").(string),
		ChartName:    d.Get("chart_name").(string),
		ChartVersion: d.Get("chart_version").(string),
		Namespace:    d.Get("namespace").(string),
		Overrides:    d.Get("overrides").(string),
	}
	if err := cmd.Delete(release); err != nil {
		return err
	}
	d.SetId("")
	return nil
}

func resourceManifestRead(d *schema.ResourceData, m interface{}) error {
	cmd := m.(*helmcmd.HelmCmd)

	release := &helmcmd.HelmRelease{
		Name: d.Id(),
	}

	if err := cmd.Read(release); err != nil {
		// TODO: Handle the case properly where the resource does not exist,
		// not with error, but with setting the ID to ""
		if err == helmcmd.ErrHelmNotExist {
			d.SetId("")
			return nil
		} else {
			return err
		}
	}

	d.Set("name", release.Name)
	d.Set("chart_name", release.ChartName)
	d.Set("chart_version", release.ChartVersion)
	d.Set("namespace", release.Namespace)
	d.Set("overrides", release.Overrides)

	return nil
}
