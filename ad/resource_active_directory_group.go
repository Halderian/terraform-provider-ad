package ad

import (
	"fmt"
	"log"
	"strings"

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
			"domain": {
				Type:        schema.TypeString,
				Description: "The domain of the group",
				Required:    true,
				ForceNew:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "The description of the group",
				Optional:    true,
				Default:     nil,
				ForceNew:    false,
			},
			"orgunit": {
				Type:        schema.TypeString,
				Description: "The organizational unit the group belongs to",
				Optional:    true,
				Default:     nil,
				ForceNew:    false,
			},
			"type": {
				Type:        schema.TypeString,
				Description: "The type of the group. Could be either GLOBAL or LOCAL. Defaults to GLOBAL.",
				Optional:    true,
				Default:     "GLOBAL",
				ForceNew:    true,
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
	domain := d.Get("domain").(string)
	orgunit := d.Get("orgunit").(string)
	description := d.Get("description").(string)
	typeOfGroup := d.Get("type").(string)

	dnOfGroup := "cn=" + groupName

	if orgunit != "" {
		dnOfGroup += "," + orgunit
	} else {
		dnOfGroup += ",cn=Users"
		domainArr := strings.Split(domain, ".")
		for _, item := range domainArr {
			dnOfGroup += ",dc=" + item
		}
	}

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
	return resourceADGroupRead(d, meta)
}

func resourceADGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	//TODO
	return resourceADGroupRead(d, meta)
}

func resourceADGroupDelete(d *schema.ResourceData, meta interface{}) error {
	groupName := d.Get("name").(string)
	domain := d.Get("domain").(string)
	orgunit := d.Get("orgunit").(string)
	dnOfGroup := "cn=" + groupName

	if orgunit != "" {
		dnOfGroup += "," + orgunit
	} else {
		dnOfGroup += ",cn=Users"
		domainArr := strings.Split(domain, ".")
		for _, item := range domainArr {
			dnOfGroup += ",dc=" + item
		}
	}

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
	dnOfGroup := d.Get("dn").(string)

	if dnOfGroup == "" {
		groupName = d.Get("name").(string)
		domain := d.Get("domain").(string)
		orgunit := d.Get("orgunit").(string)
		dnOfGroup = "cn=" + groupName

		if orgunit != "" {
			dnOfGroup += "," + orgunit
		} else {
			dnOfGroup += ",cn=Users"
			domainArr := strings.Split(domain, ".")
			for _, item := range domainArr {
				dnOfGroup += ",dc=" + item
			}
		}
	} else {
		groupName = parseDN(dnOfGroup, "cn")
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
		d.SetId(groupID)
		d.Set("dn", groupDN)
		d.Set("name", group.GetAttributeValue("cn"))
		d.Set("description", group.GetAttributeValue("description"))
	}
	return nil
}
