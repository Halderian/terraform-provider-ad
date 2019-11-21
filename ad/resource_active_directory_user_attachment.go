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
    Schema: map[string]*schema.Schema{
      "group_dn": {
        Type:         schema.TypeString,
        Description:  "The dn of the group to add the user to.",
        Required:     true,
        ForceNew:     false,
      },
      "user_dn": {
        Type:         schema.TypeString,
        Description:  "The dn of the user to attache to the the group.",
        Required:     true,
        ForceNew:     true,
      },
    },
  }
}

func resourceADUserAttachmentCreate(d *schema.ResourceData, meta interface{}) error {
	groupDN := d.Get("group_dn").(string)
	userDN 	:= d.Get("user_dn").(string)

	client 	:= meta.(*ldap.Conn)

	err := addMemberToGroup(groupDN, userDN, client)
	if err != nil {
		log.Printf("[ERROR] Error while attaching user to group: %s", err)
		return fmt.Errorf("Error while attaching user to group: %s", err)
	}

	d.SetId(groupDN+userDN)

  return resourceADUserAttachmentRead(d, meta)
}

func resourceADUserAttachmentUpdate(d *schema.ResourceData, meta interface{}) error {


  return resourceADUserAttachmentRead(d, meta)
}

func resourceADUserAttachmentDelete(d *schema.ResourceData, meta interface{}) error {


  return resourceADUserAttachmentRead(d, meta)
}

func resourceADUserAttachmentRead(d *schema.ResourceData, meta interface{}) error {


  return nil
}
