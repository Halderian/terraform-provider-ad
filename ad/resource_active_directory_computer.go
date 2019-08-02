package ad

import (
	"fmt"
	"log"
	"strings"

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
				ForceNew:    true,
			},
			"domain": {
				Type:        schema.TypeString,
				Description: "The domain of the computer",
				Required:    true,
				ForceNew:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "The description of the computer",
				Optional:    true,
				Default:     nil,
				ForceNew:    false,
			},
			"orgunit": {
				Type:        schema.TypeString,
				Description: "The organizational unit the computer belongs to",
				Optional:    true,
				Default:     nil,
				ForceNew:    false,
			},
		},
	}
}

func resourceADComputerCreate(d *schema.ResourceData, meta interface{}) error {
	computerName := d.Get("name").(string)
	domain := d.Get("domain").(string)
	orgunit := d.Get("orgunit").(string)
	description := d.Get("description").(string)

	dnOfComputer := "cn=" + computerName
	if orgunit != "" {
		dnOfComputer += ",ou=Computers,ou=" + orgunit
	} else {
		dnOfComputer += ",cn=Computers"
	}
	domainArr := strings.Split(domain, ".")
	for _, item := range domainArr {
		dnOfComputer += ",dc=" + item
	}

	log.Printf("[DEBUG] Name of the DN is: %s", dnOfComputer)
	log.Printf("[DEBUG] Adding the computer to the AD: %s", computerName)

	client := meta.(*ldap.Conn)

	err := addComputerToAD(computerName, dnOfComputer, client, description)
	if err != nil {
		log.Printf("[ERROR] Error while adding a computer to the AD: %s ", err)
		return fmt.Errorf("Error while adding a computer to the AD %s", err)
	}
	log.Printf("[DEBUG] Computer added to AD successfully: %s", computerName)
	d.SetId(dnOfComputer)
	return resourceADComputerRead(d, meta)
}

func resourceADComputerUpdate(d *schema.ResourceData, meta interface{}) error {
	//TODO
	return resourceADComputerRead(d, meta)
}

func resourceADComputerDelete(d *schema.ResourceData, meta interface{}) error {
	computerName := d.Get("name").(string)
	domain := d.Get("domain").(string)
	orgunit := d.Get("orgunit").(string)

	dnOfComputer := "cn=" + computerName
	if orgunit != "" {
		dnOfComputer += ",ou=Computers,ou=" + orgunit
	} else {
		dnOfComputer += ",cn=Computers"
	}
	domainArr := strings.Split(domain, ".")
	for _, item := range domainArr {
		dnOfComputer += ",dc=" + item
	}

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
	computerName := d.Get("name").(string)
	domain := d.Get("domain").(string)
	orgunit := d.Get("orgunit").(string)

	var dnOfComputer string
	if orgunit != "" {
		dnOfComputer += "ou=Computers,ou=" + orgunit
	} else {
		dnOfComputer += "cn=Computers"
	}
	domainArr := strings.Split(domain, ".")
	for _, item := range domainArr {
		dnOfComputer += ",dc=" + item
	}

	log.Printf("[DEBUG] Name of the DN is: %s", dnOfComputer)
	log.Printf("[DEBUG] Refreshing the computer from the AD: %s", computerName)

	client := meta.(*ldap.Conn)

	searchRequest := ldap.NewSearchRequest(
		dnOfComputer, // The base dn to search
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(&(objectClass=Computer)(cn="+computerName+"))", // The filter to apply
		[]string{"dn", "cn", "description"},              // A list attributes to retrieve
		nil,
	)

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
		d.SetId(computer.DN)
		d.Set("name", computer.GetAttributeValue("cn"))
		d.Set("description", computer.GetAttributeValue("description"))
	}
	return nil
}
