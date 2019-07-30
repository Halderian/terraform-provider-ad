package ad

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	ldap "gopkg.in/ldap.v2"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceUser() *schema.Resource {
	return &schema.Resource{
		Create: resourceADUserCreate,
		Read:   resourceADUserRead,
		Delete: resourceADUserDelete,
		Schema: map[string]*schema.Schema{
			"username": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"password": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
				ForceNew:  true,
			},
			"domain": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"firstname": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"lastname": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  nil,
				ForceNew: true,
			},
		},
	}
}

func resourceADUserCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ldap.Conn)

	user := d.Get("username").(string)
	password := d.Get("password").(string)
	domain := d.Get("domain").(string)
	description := d.Get("description").(string)
	firstname := d.Get("firstname").(string)
	lastname := d.Get("lastname").(string)
	var dnOfUser string
	dnOfUser += "CN=" + firstname + " " + lastname
	domainArr := strings.Split(domain, ".")
	dnOfUser += ",OU=Users,OU=" + domainArr[0]
	for _, item := range domainArr {
		dnOfUser += ",DC=" + item
	}

	log.Printf("[DEBUG] Name of the DN is : %s ", dnOfUser)
	log.Printf("[DEBUG] Adding the User to the AD : %s ", user)

	err := addUserToAD(user, password, firstname, lastname, dnOfUser, client, description)
	if err != nil {
		log.Printf("[ERROR] Error while adding a User to the AD : %s ", err)
		return fmt.Errorf("Error while adding a User to the AD %s", err)
	}
	log.Printf("[DEBUG] User Added to AD successfully: %s", user)
	d.SetId(domain + "/" + user)
	return nil
}

func resourceADUserRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ldap.Conn)

	user := d.Get("username").(string)
	domain := d.Get("domain").(string)
	var dnOfUser string
	domainArr := strings.Split(domain, ".")
	dnOfUser = "dc=" + domainArr[0]
	for index, item := range domainArr {
		if index == 0 {
			continue
		}
		dnOfUser += ",dc=" + item
	}
	log.Printf("[DEBUG] Name of the DN is : %s ", dnOfUser)
	log.Printf("[DEBUG] Searching the User in the AD : %s ", user)

	searchRequest := ldap.NewSearchRequest(
		dnOfUser, // The base dn to search
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(&(objectClass=User)(cn="+user+"))", // The filter to apply
		[]string{"dn", "cn"},                 // A list attributes to retrieve
		nil,
	)

	sr, err := client.Search(searchRequest)
	if err != nil {
		log.Printf("[ERROR] Error while searching a User : %s ", err)
		return fmt.Errorf("Error while searching a User : %s", err)
	}
	fmt.Println("[ERROR] Found " + strconv.Itoa(len(sr.Entries)) + " Entries")
	for _, entry := range sr.Entries {
		fmt.Printf("%s: %v\n", entry.DN, entry.GetAttributeValue("cn"))
	}
	if len(sr.Entries) == 0 {
		log.Println("[ERROR] User was not found")
		d.SetId("")
	}
	return nil
}

func resourceADUserDelete(d *schema.ResourceData, meta interface{}) error {
	resourceADUserRead(d, meta)
	if d.Id() == "" {
		log.Println("[ERROR] Cannot find User in the specified AD")
		return fmt.Errorf("[ERROR] Cannot find User in the specified AD")
	}
	client := meta.(*ldap.Conn)

	user := d.Get("username").(string)
	domain := d.Get("domain").(string)
	var dnOfUser string
	dnOfUser += "cn=" + user + ",ou=User"
	domainArr := strings.Split(domain, ".")
	for _, item := range domainArr {
		dnOfUser += ",dc=" + item
	}

	log.Printf("[DEBUG] Name of the DN is : %s ", dnOfUser)
	log.Printf("[DEBUG] Deleting the User from the AD : %s ", user)

	err := deleteUserFromAD(dnOfUser, client)
	if err != nil {
		log.Printf("[ERROR] Error while Deleting a User from AD : %s ", err)
		return fmt.Errorf("Error while Deleting a User from AD %s", err)
	}
	log.Printf("[DEBUG] User deleted from AD successfully: %s", user)
	return nil
}
