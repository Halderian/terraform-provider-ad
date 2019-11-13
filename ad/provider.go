package ad

import (
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"domain": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The Domain in which AD Server resides",
				DefaultFunc: schema.EnvDefaultFunc("AD_DOMAIN", nil),
			},

			"ip": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The IP of the AD Server",
				DefaultFunc: schema.EnvDefaultFunc("AD_IP", nil),
			},

			"user": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The user name of the AD Server",
				DefaultFunc: schema.EnvDefaultFunc("AD_USER", nil),
			},

			"password": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "The user password of the AD Server",
				DefaultFunc: schema.EnvDefaultFunc("AD_PASSWORD", nil),
			},

			"ssl": {
				Type:        schema.TypeBool,
				Required:    true,
				Description: "Use an SSL connection to the AD",
				DefaultFunc: schema.EnvDefaultFunc("AD_SSL", true),
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"ad_computer": resourceComputer(),
			"ad_group":    resourceGroup(),
			"ad_ou":       resourceOrgUnit(),
			"ad_user":     resourceUser(),
		},

		DataSourcesMap: map[string]*schema.Resource{
			"ad_group": dataActiveDirectoryGroup(),
			"ad_ou":    dataActiveDirectoryOrgUnit(),
			"ad_user":  dataActiveDirectoryUser(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {

	config := Config{
		Domain:   d.Get("domain").(string),
		IP:       d.Get("ip").(string),
		Username: d.Get("user").(string),
		Password: d.Get("password").(string),
		UseSSL:   d.Get("ssl").(bool),
	}
	log.Printf("[DEBUG] Connecting to AD")
	return config.Client()
}

func expandStringSlice(configured []interface{}) []string {
	vs := make([]string, 0, len(configured))
	for _, v := range configured {
		val, ok := v.(string)
		if ok && val != "" {
			vs = append(vs, v.(string))
		}
	}
	return vs
}
