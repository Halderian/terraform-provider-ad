package ad

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func dataActiveDirectoryGroup() *schema.Resource {
	return &schema.Resource{
		Read: resourceADGroupRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the group",
				Required:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "The description of the group",
				Computed:    true,
			},
			"parent": {
				Type:        schema.TypeString,
				Description: "The parent the group belongs to. Could be either the DN of an OU or a DC.",
				Optional:    true,
				Default:     nil,
			},
			"type": {
				Type:        schema.TypeString,
				Description: "The type of the group. Could be either GLOBAL or LOCAL. Defaults to GLOBAL.",
				Optional:    true,
				Default:     "GLOBAL",
			},
			"dn": {
				Type:        schema.TypeString,
				Description: "The distinguished name of the group",
				Optional:    true,
				Default:     nil,
			},
			"members": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},
		},
	}
}
