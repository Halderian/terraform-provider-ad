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
			"domain": {
				Type:        schema.TypeString,
				Description: "The domain of the group",
				Optional:    true,
				Default:     nil,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "The description of the group",
				Computed:    true,
			},
			"orgunit": {
				Type:        schema.TypeString,
				Description: "The organizational unit the group belongs to",
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
		},
	}
}
