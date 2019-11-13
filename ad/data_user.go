package ad

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func dataActiveDirectoryUser() *schema.Resource {
	return &schema.Resource{
		Read: resourceADUserRead,
		Schema: map[string]*schema.Schema{
			"username": {
				Type:        schema.TypeString,
				Description: "The login name of the user",
				Required:    true,
			},
			"password": {
				Type:        schema.TypeString,
				Description: "The login password of the user",
				Computed:    true,
				Sensitive:   true,
			},
			"domain": {
				Type:        schema.TypeString,
				Description: "The login domain of the user",
				Required:    true,
			},
			"firstname": {
				Type:        schema.TypeString,
				Description: "The first name of the user",
				Computed:    true,
			},
			"lastname": {
				Type:        schema.TypeString,
				Description: "The last name of the user",
				Computed:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "The full name of the user",
				Computed:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "The description of the user",
				Computed:    true,
			},
			"orgunit": {
				Type:        schema.TypeString,
				Description: "The organizational unit the user belongs to",
				Optional:    true,
				Default:     nil,
			},
			"dn": {
				Type:        schema.TypeString,
				Description: "The distinguished name of the user",
				Computed:    true,
			},
		},
	}
}
