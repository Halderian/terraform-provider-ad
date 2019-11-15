package ad

import (
	"fmt"
	"log"

	ldap "gopkg.in/ldap.v3"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceADGroupCreate,
		Read:   resourceADGroupRead,
		Update: resourceADGroupUpdate,
		Delete: resourceADGroupDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the group",
				Required:    true,
				ForceNew:    false,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "The description of the group",
				Optional:    true,
				Default:     nil,
				ForceNew:    false,
			},
			"parent": {
				Type:        schema.TypeString,
				Description: "The parent the group belongs to. Could be either the DN of an OU or a DC.",
				Required:    true,
				ForceNew:    false,
			},
			"type": {
				Type:        schema.TypeString,
				Description: "The type of the group. Could be either GLOBAL or LOCAL. Defaults to GLOBAL.",
				Optional:    true,
				Default:     "GLOBAL",
				ForceNew:    true,
			},
			"members": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
				Computed: true,
				ForceNew: false,
			},
			"dn": {
				Type:        schema.TypeString,
				Description: "The distinguished name of the group",
				Computed:    true,
			},
		},
	}
}

func resourceADGroupCreate(d *schema.ResourceData, meta interface{}) error {
	groupName := d.Get("name").(string)
	parent := d.Get("parent").(string)
	description := d.Get("description").(string)
	typeOfGroup := d.Get("type").(string)

	dnOfGroup := fmt.Sprintf("cn=%s,%s", groupName, parent)

	log.Printf("[DEBUG] Name of the DN is : %s", dnOfGroup)
	log.Printf("[DEBUG] Adding the group to the AD: %s ", groupName)

	client := meta.(*ldap.Conn)

	err := addGroupToAD(groupName, dnOfGroup, typeOfGroup, client, description)
	if err != nil {
		log.Printf("[ERROR] Error while adding a group to the AD : %s", err)
		return fmt.Errorf("Error while adding a group to the AD %s", err)
	}
	log.Printf("[DEBUG] Group added to AD successfully: %s", groupName)
	d.Set("dn", dnOfGroup)

	for _, v := range d.Get("members").(*schema.Set).List() {
		log.Printf("[DEBUG] Found new member %s", v)
		err = addMemberToGroup(dnOfGroup, v.(string), client)
		if err != nil {
			log.Printf("[ERROR] Error while adding a member to the group : %s", err)
			return fmt.Errorf("Error while adding a member to the group %s", err)
		}
	}

	return resourceADGroupRead(d, meta)
}

func resourceADGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	groupName := d.Get("name").(string)
	parent := d.Get("parent").(string)

	var dnOfGroup string
	var err error
	client := meta.(*ldap.Conn)

	if d.HasChange("parent") || d.HasChange("name") {
		origName := groupName
		origParent := parent

		if d.HasChange("parent") {
			o, _ := d.GetChange("parent")
			origParent = o.(string)
		}

		if d.HasChange("name") {
			o, _ := d.GetChange("name")
			origName = o.(string)
		}

		if err == nil && origName != groupName {
			// first: rename group
			dnOfGroup = fmt.Sprintf("cn=%s,%s", origName, origParent)
			log.Printf("[DEBUG] Name of the DN is : %s", dnOfGroup)
			log.Printf("[DEBUG] About to rename the group to %s", groupName)
			err = renameADEntry(dnOfGroup, fmt.Sprintf("cn=%s", groupName), client)
		}

		if err == nil && origParent != parent {
			// next: move group to new parent
			dnOfGroup = fmt.Sprintf("cn=%s,%s", groupName, origParent)
			log.Printf("[DEBUG] Name of the DN is : %s", dnOfGroup)
			log.Printf("[DEBUG] About to move the group to %s", parent)
			err = moveADEntry(dnOfGroup, parent, client)
		}
	}

	dnOfGroup = fmt.Sprintf("cn=%s,%s", groupName, parent)

	log.Printf("[DEBUG] Name of the DN is : %s", dnOfGroup)

	if err == nil && d.HasChange("description") {
		new := d.Get("description").(string)
		log.Printf("[DEBUG] found new description %s. Do update", new)
		err = updateADEntry(dnOfGroup, "description", new, client)
	}

	if err == nil && d.HasChange("members") {
		old, new := d.GetChange("members")
		oldList := old.(*schema.Set).List()
		newList := new.(*schema.Set).List()
		for _, v := range newList {
			if itemExists(oldList, v) {
				log.Printf("[DEBUG] found existing member %s. Skip update", v)
			} else {
				log.Printf("[DEBUG] found new member %s. Do update (add)", v)
				err = addMemberToGroup(dnOfGroup, v.(string), client)
			}
		}
		for _, v := range oldList {
			if itemExists(newList, v) {
				log.Printf("[DEBUG] found existing member %s. Skip update", v)
			} else {
				log.Printf("[DEBUG] found obsolete member %s. Do update (remove)", v)
				err = removeMemberFromGroup(dnOfGroup, v.(string), client)
			}
		}
	}

	if err != nil {
		log.Printf("[ERROR] Error while modifying a group from AD : %s ", err)
		return fmt.Errorf("Error while modifying a group from AD %s", err)
	}

	d.Set("dn", dnOfGroup)
	return resourceADGroupRead(d, meta)
}

func resourceADGroupDelete(d *schema.ResourceData, meta interface{}) error {
	groupName := d.Get("name").(string)
	parent := d.Get("parent").(string)

	dnOfGroup := fmt.Sprintf("cn=%s,%s", groupName, parent)

	log.Printf("[DEBUG] Name of the DN is : %s", dnOfGroup)
	log.Printf("[DEBUG] Deleting the group from the AD : %s", groupName)

	resourceADGroupRead(d, meta)
	if d.Id() == "" {
		log.Printf("[DEBUG] Group has been already removed from AD: %s", groupName)
		return nil
	}

	client := meta.(*ldap.Conn)

	err := deleteGroupFromAD(dnOfGroup, client)
	if err != nil {
		log.Printf("[ERROR] Error while deleting a group from AD : %s ", err)
		return fmt.Errorf("Error while deleting a group from AD %s", err)
	}
	log.Printf("[DEBUG] Group deleted from AD successfully: %s", groupName)
	return nil
}

func resourceADGroupRead(d *schema.ResourceData, meta interface{}) error {
	var groupName string
	var parent string

	dnOfGroup := d.Get("dn").(string)

	if dnOfGroup == "" {
		groupName = d.Get("name").(string)
		parent = d.Get("parent").(string)

		dnOfGroup = fmt.Sprintf("cn=%s,%s", groupName, parent)
	} else {
		groupName, parent = parseDN(dnOfGroup, "cn")
	}

	log.Printf("[DEBUG] Name of the DN is : %s ", dnOfGroup)
	log.Printf("[DEBUG] Searching the group from the AD : %s ", groupName)

	client := meta.(*ldap.Conn)

	searchParam := "(distinguishedName=" + dnOfGroup + ")"

	if d.Id() != "" {
		searchParam = "(objectGUID=" + generateObjectIdQueryString(d.Id()) + ")"
	}

	log.Printf("[DEBUG] Search Parameters for group: %s ", searchParam)

	searchRequest := ldap.NewSearchRequest(
		dnOfGroup, // The base dn to search
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(&(objectClass=group)"+searchParam+")", // The filter to apply
		[]string{"dn", "cn", "description"},     // A list attributes to retrieve
		nil,
	)

	searchRequest.Controls = append(searchRequest.Controls, &ldapControlServerExtendDN{})

	sr, err := client.Search(searchRequest)
	if err != nil {
		log.Printf("[ERROR] Error while searching a group: %s", err)
		return fmt.Errorf("Error while searching a group: %s", err)
	}
	if len(sr.Entries) == 0 {
		log.Println("[ERROR] Group was not found")
		d.SetId("")
	} else {
		if len(sr.Entries) > 1 {
			log.Printf("[ERROR] Error found ambigious values for group: %s", groupName)
			return fmt.Errorf("Error found ambigious values for group: %s", groupName)
		}
		group := sr.Entries[0]
		groupID, groupDN := parseExtendedDN(group.DN)
		groupName, parent = parseDN(groupDN, "cn")
		d.SetId(groupID)
		d.Set("dn", groupDN)
		d.Set("name", groupName)
		d.Set("description", group.GetAttributeValue("description"))
		d.Set("parent", parent)
	}
	return nil
}
