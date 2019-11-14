package ad

import (
	"fmt"
	"log"

	ldap "gopkg.in/ldap.v3"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceOrgUnit() *schema.Resource {
	return &schema.Resource{
		Create: resourceADOrgUnitCreate,
		Read:   resourceADOrgUnitRead,
		Update: resourceADOrgUnitUpdate,
		Delete: resourceADOrgUnitDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the organizational unit",
				Required:    true,
				ForceNew:    false,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "The description of the organizational unit",
				Optional:    true,
				Default:     nil,
				ForceNew:    false,
			},
			"parent": {
				Type:        schema.TypeString,
				Description: "The parent the organizational unit belongs to. Could be either the DN of an OU or a DC.",
				Required:    true,
				ForceNew:    false,
			},
			"dn": {
				Type:        schema.TypeString,
				Description: "The distinguished name of the organization unit",
				Computed:    true,
			},
		},
	}
}

func resourceADOrgUnitCreate(d *schema.ResourceData, meta interface{}) error {
	orgUnitName := d.Get("name").(string)
	parent := d.Get("parent").(string)
	description := d.Get("description").(string)

	dnOfOrgUnit := fmt.Sprintf("ou=%s,%s", orgUnitName, parent)

	log.Printf("[DEBUG] Name of the DN is : %s", dnOfOrgUnit)
	log.Printf("[DEBUG] Adding the organizational unit to the AD: %s ", orgUnitName)

	client := meta.(*ldap.Conn)

	err := addOrgUnitToAD(orgUnitName, dnOfOrgUnit, client, description)
	if err != nil {
		log.Printf("[ERROR] Error while adding a organizational unit to the AD : %s", err)
		return fmt.Errorf("Error while adding a organizational unit to the AD %s", err)
	}
	log.Printf("[DEBUG] Organizational Unit added to AD successfully: %s", orgUnitName)
	d.Set("dn", dnOfOrgUnit)
	return resourceADOrgUnitRead(d, meta)
}

func resourceADOrgUnitUpdate(d *schema.ResourceData, meta interface{}) error {
	//TODO
	return resourceADOrgUnitRead(d, meta)
}

func resourceADOrgUnitDelete(d *schema.ResourceData, meta interface{}) error {
	orgUnitName := d.Get("name").(string)
	parent := d.Get("parent").(string)

	dnOfOrgUnit := fmt.Sprintf("ou=%s,%s", orgUnitName, parent)

	log.Printf("[DEBUG] Name of the DN is : %s", dnOfOrgUnit)
	log.Printf("[DEBUG] Deleting the organizational unit from the AD : %s", orgUnitName)

	resourceADOrgUnitRead(d, meta)
	if d.Id() == "" {
		log.Printf("[DEBUG] Organizational Unit has been already removed from AD: %s", orgUnitName)
		return nil
	}

	client := meta.(*ldap.Conn)

	err := deleteOrgUnitFromAD(dnOfOrgUnit, client)
	if err != nil {
		log.Printf("[ERROR] Error while deleting a organizational unit from AD : %s ", err)
		return fmt.Errorf("Error while deleting a organizational unit from AD %s", err)
	}
	log.Printf("[DEBUG] Organizational Unit deleted from AD successfully: %s", orgUnitName)
	return nil
}

func resourceADOrgUnitRead(d *schema.ResourceData, meta interface{}) error {
	var orgUnitName string
	var parent string

	dnOfOrgUnit := d.Get("dn").(string)

	if dnOfOrgUnit == "" {
		orgUnitName = d.Get("name").(string)
		parent = d.Get("parent").(string)

		dnOfOrgUnit = fmt.Sprintf("ou=%s,%s", orgUnitName, parent)
	} else {
		orgUnitName, parent = parseDN(dnOfOrgUnit, "ou")
	}

	log.Printf("[DEBUG] Name of the DN is : %s ", dnOfOrgUnit)
	log.Printf("[DEBUG] Searching the organizational unit from the AD : %s ", orgUnitName)

	client := meta.(*ldap.Conn)

	searchParam := "(distinguishedName=" + dnOfOrgUnit + ")"

	if d.Id() != "" {
		searchParam = "(objectGUID=" + generateObjectIdQueryString(d.Id()) + ")"
	}

	log.Printf("[DEBUG] Search Parameters for organizational unit: %s ", searchParam)

	searchRequest := ldap.NewSearchRequest(
		dnOfOrgUnit, // The base dn to search
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(&(objectClass=organizationalunit)"+searchParam+")", // The filter to apply
		[]string{"dn", "ou", "description"},                  // A list attributes to retrieve
		nil,
	)

	searchRequest.Controls = append(searchRequest.Controls, &ldapControlServerExtendDN{})

	sr, err := client.Search(searchRequest)
	if err != nil {
		log.Printf("[ERROR] Error while searching a organizational unit: %s", err)
		return fmt.Errorf("Error while searching a organizational unit: %s", err)
	}
	if len(sr.Entries) == 0 {
		log.Println("[ERROR] Organizational Unit was not found")
		d.SetId("")
	} else {
		if len(sr.Entries) > 1 {
			log.Printf("[ERROR] Error found ambigious values for organizational unit: %s", orgUnitName)
			return fmt.Errorf("Error found ambigious values for organizational unit: %s", orgUnitName)
		}
		orgUnit := sr.Entries[0]
		orgID, orgDN := parseExtendedDN(orgUnit.DN)
		orgUnitName, parent = parseDN(orgDN, "ou")
		d.SetId(orgID)
		d.Set("dn", orgDN)
		d.Set("name", orgUnitName)
		d.Set("description", orgUnit.GetAttributeValue("description"))
		d.Set("parent", parent)
	}
	return nil
}
