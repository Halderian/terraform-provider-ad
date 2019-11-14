package ad

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func dataActiveDirectoryOrgUnit() *schema.Resource {
	return &schema.Resource{
		Read: resourceADOrgUnitRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the organizational unit",
				Required:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "The description of the organizational unit",
				Computed:    true,
			},
			"parent": {
				Type:        schema.TypeString,
				Description: "The parent the organizational unit belongs to. Could be either the DN of an OU or a DC.",
				Optional:    true,
				Default:     nil,
			},
			"dn": {
				Type:        schema.TypeString,
				Description: "The distinguished name of the organization unit",
				Optional:    true,
				Default:     nil,
			},
		},
	}
}
