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
			"parent": {
				Type:        schema.TypeString,
				Description: "The parent the domain belongs to. Could be either the DN of an OU or a DC.",
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
			"dn": {
				Type:        schema.TypeString,
				Description: "The distinguished name of the user",
				Computed:    true,
			},
			"groups": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},
		},
	}
}
