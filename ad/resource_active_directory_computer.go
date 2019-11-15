package ad

import (
	"fmt"
	"log"

	ldap "gopkg.in/ldap.v3"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceComputer() *schema.Resource {
	return &schema.Resource{
		Create: resourceADComputerCreate,
		Read:   resourceADComputerRead,
		Update: resourceADComputerUpdate,
		Delete: resourceADComputerDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the computer",
				Required:    true,
				ForceNew:    false,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "The description of the computer",
				Optional:    true,
				Default:     nil,
				ForceNew:    false,
			},
			"parent": {
				Type:        schema.TypeString,
				Description: "The parent the computer belongs to. Could be either the DN of an OU or a DC.",
				Required:    true,
				ForceNew:    false,
			},
			"dn": {
				Type:        schema.TypeString,
				Description: "The distinguished name of the computer",
				Computed:    true,
			},
		},
	}
}

func resourceADComputerCreate(d *schema.ResourceData, meta interface{}) error {
	computerName := d.Get("name").(string)
	parent := d.Get("parent").(string)
	description := d.Get("description").(string)

	dnOfComputer := fmt.Sprintf("cn=%s,%s", computerName, parent)

	log.Printf("[DEBUG] Name of the DN is: %s", dnOfComputer)
	log.Printf("[DEBUG] Adding the computer to the AD: %s", computerName)

	client := meta.(*ldap.Conn)

	err := addComputerToAD(computerName, dnOfComputer, client, description)
	if err != nil {
		log.Printf("[ERROR] Error while adding a computer to the AD: %s ", err)
		return fmt.Errorf("Error while adding a computer to the AD %s", err)
	}
	log.Printf("[DEBUG] Computer added to AD successfully: %s", computerName)
	d.Set("dn", dnOfComputer)
	return resourceADComputerRead(d, meta)
}

func resourceADComputerUpdate(d *schema.ResourceData, meta interface{}) error {
	//TODO
	return resourceADComputerRead(d, meta)
}

func resourceADComputerDelete(d *schema.ResourceData, meta interface{}) error {
	computerName := d.Get("name").(string)
	parent := d.Get("parent").(string)

	dnOfComputer := fmt.Sprintf("cn=%s,%s", computerName, parent)

	log.Printf("[DEBUG] Name of the DN is: %s", dnOfComputer)
	log.Printf("[DEBUG] Deleting computer from the AD: %s", computerName)

	resourceADComputerRead(d, meta)
	if d.Id() == "" {
		log.Printf("[DEBUG] Computer has been already removed from AD: %s", computerName)
		return nil
	}

	client := meta.(*ldap.Conn)

	err := deleteComputerFromAD(dnOfComputer, client)
	if err != nil {
		log.Printf("[ERROR] Error while deleting computer from AD: %s", err)
		return fmt.Errorf("Error while deleting computer from AD %s", err)
	}
	log.Printf("[DEBUG] Computer deleted from AD successfully: %s", computerName)
	return nil
}

func resourceADComputerRead(d *schema.ResourceData, meta interface{}) error {
	var computerName string
	var parent string

	dnOfComputer := d.Get("dn").(string)

	if dnOfComputer == "" {
		computerName = d.Get("name").(string)
		parent = d.Get("parent").(string)

		dnOfComputer = fmt.Sprintf("cn=%s,%s", computerName, parent)
	} else {
		computerName, parent = parseDN(dnOfComputer, "cn")
	}

	log.Printf("[DEBUG] Name of the DN is: %s", dnOfComputer)
	log.Printf("[DEBUG] Refreshing the computer from the AD: %s", computerName)

	client := meta.(*ldap.Conn)

	searchParam := "(distinguishedName=" + dnOfComputer + ")"
	_, searchBaseDN := parseDN(dnOfComputer, "cn")

	if d.Id() != "" {
		searchParam = "(objectGUID=" + generateObjectIdQueryString(d.Id()) + ")"
		searchBaseDN = extractDomainFromDN(dnOfComputer)
	}

	log.Printf("[DEBUG] Search Parameters for computer: %s ", searchParam)

	searchRequest := ldap.NewSearchRequest(
		searchBaseDN, // The base dn to search
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(&(objectClass=Computer)"+searchParam+")", // The filter to apply
		[]string{"dn", "cn", "description"},        // A list attributes to retrieve
		nil,
	)

	searchRequest.Controls = append(searchRequest.Controls, &ldapControlServerExtendDN{})

	sr, err := client.Search(searchRequest)
	if err != nil {
		log.Printf("[ERROR] Error while searching a computer: %s", err)
		return fmt.Errorf("Error while searching a computer: %s", err)
	}
	if len(sr.Entries) == 0 {
		log.Println("[ERROR] computer was not found")
		d.SetId("")
	} else {
		if len(sr.Entries) > 1 {
			log.Printf("[ERROR] Error found ambigious values for computer: %s", computerName)
			return fmt.Errorf("Error found ambigious values for computer: %s", computerName)
		}
		computer := sr.Entries[0]
		computerID, computerDN := parseExtendedDN(computer.DN)
		computerName, parent = parseDN(computerDN, "cn")
		d.SetId(computerID)
		d.Set("dn", computerDN)
		d.Set("name", computerName)
		d.Set("description", computer.GetAttributeValue("description"))
		d.Set("parent", parent)
	}
	return nil
}
