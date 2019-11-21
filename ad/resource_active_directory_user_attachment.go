package ad

import (
	"fmt"
	"log"

	ldap "gopkg.in/ldap.v3"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceUserAttachment() *schema.Resource {
  return &schema.Resource{
    Create: resourceADUserAttachmentCreate,
    Read:   resourceADUserAttachmentRead,
    Update: resourceADUserAttachmentUpdate,
    Delete: resourceADUserAttachmentDelete,
    Schema: map[String]*schema.Schema{
      "group_name": {
        Type:         schema.TypeString,
        Description:  "The name of the group to add the user to.",
        Required:     true,
        ForceNew:     false,
      },
      "user_name": {
        Type:         schema.TypeString,
        Description:  "The name of the user to attache to the the group.",
        Required:     true,
        ForceNew:     true,
      },
    }
  }
}

func resourceADUserAttachmentCreate(d *schema.ResourceData, meta interface{}) error {


  return resourceADUserAttachmentRead(d, meta)
}

func resourceADUserAttachmentUpdate(d *schema.ResourceData, meta interface{}) error {


  return resourceADUserAttachmentRead(d, meta)
}

func resourceADUserAttachmentDelete(d *schema.ResourceData, meta interface{}) error {


  return resourceADUserAttachmentRead(d, meta)
}
