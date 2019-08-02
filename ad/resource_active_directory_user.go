package ad

import (
	"fmt"
	"log"
	"strings"

	ldap "gopkg.in/ldap.v3"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceUser() *schema.Resource {
	return &schema.Resource{
		Create: resourceADUserCreate,
		Read:   resourceADUserRead,
		Update: resourceADUserUpdate,
		Delete: resourceADUserDelete,
		Schema: map[string]*schema.Schema{
			"username": {
				Type:        schema.TypeString,
				Description: "The login name of the user",
				Required:    true,
				ForceNew:    true,
			},
			"password": {
				Type:        schema.TypeString,
				Description: "The login password of the user",
				Required:    true,
				Sensitive:   true,
				ForceNew:    false,
			},
			"domain": {
				Type:        schema.TypeString,
				Description: "The login domain of the user",
				Required:    true,
				ForceNew:    true,
			},
			"firstname": {
				Type:        schema.TypeString,
				Description: "The first name of the user",
				Required:    true,
				ForceNew:    false,
			},
			"lastname": {
				Type:        schema.TypeString,
				Description: "The last name of the user",
				Required:    true,
				ForceNew:    false,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "The full name of the user",
				Computed:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "The description of the user",
				Optional:    true,
				Default:     nil,
				ForceNew:    false,
			},
			"orgunit": {
				Type:        schema.TypeString,
				Description: "The organizational unit the user belongs to",
				Optional:    true,
				Default:     nil,
				ForceNew:    false,
			},
		},
	}
}

func resourceADUserCreate(d *schema.ResourceData, meta interface{}) error {
	username := d.Get("username").(string)
	password := d.Get("password").(string)
	domain := d.Get("domain").(string)
	orgunit := d.Get("orgunit").(string)
	description := d.Get("description").(string)
	firstname := d.Get("firstname").(string)
	lastname := d.Get("lastname").(string)
	name := fmt.Sprintf("%s %s", firstname, lastname)

	dnOfUser := "CN=" + name
	if orgunit != "" {
		dnOfUser += ",ou=Users,ou=" + orgunit
	} else {
		dnOfUser += ",cn=Users"
	}
	domainArr := strings.Split(domain, ".")
	for _, item := range domainArr {
		dnOfUser += ",dc=" + item
	}

	log.Printf("[DEBUG] Name of the DN is : %s", dnOfUser)
	log.Printf("[DEBUG] Adding the user to the AD : %s", name)

	client := meta.(*ldap.Conn)

	err := addUserToAD(username, firstname, lastname, dnOfUser, client, description)
	if err != nil {
		log.Printf("[ERROR] Error while adding a user to the AD : %s", err)
		return fmt.Errorf("Error while adding a user to the AD %s", err)
	}
	d.SetId(dnOfUser)
	err = setUserPassword(dnOfUser, password, client)
	if err != nil {
		log.Printf("[ERROR] Error while changing password of user : %s", err)
		return fmt.Errorf("Error while changing password of user %s", err)
	}
	err = activateUser(dnOfUser, client)
	if err != nil {
		log.Printf("[ERROR] Error while activating of user : %s", err)
		return fmt.Errorf("Error while activating of user %s", err)
	}
	log.Printf("[DEBUG] User added to AD successfully: %s", username)
	return resourceADUserRead(d, meta)
}

func resourceADUserUpdate(d *schema.ResourceData, meta interface{}) error {
	//TODO
	return resourceADUserRead(d, meta)
}

func resourceADUserDelete(d *schema.ResourceData, meta interface{}) error {
	domain := d.Get("domain").(string)
	orgunit := d.Get("orgunit").(string)
	firstname := d.Get("firstname").(string)
	lastname := d.Get("lastname").(string)
	name := fmt.Sprintf("%s %s", firstname, lastname)

	dnOfUser := "CN=" + name
	if orgunit != "" {
		dnOfUser += ",ou=Users,ou=" + orgunit
	} else {
		dnOfUser += ",cn=Users"
	}
	domainArr := strings.Split(domain, ".")
	for _, item := range domainArr {
		dnOfUser += ",dc=" + item
	}

	log.Printf("[DEBUG] Name of the DN is: %s", dnOfUser)
	log.Printf("[DEBUG] Deleting the user from the AD: %s", name)

	resourceADUserRead(d, meta)
	if d.Id() == "" {
		log.Printf("[DEBUG] User has been already removed from AD: %s", name)
		return nil
	}

	client := meta.(*ldap.Conn)

	err := deleteUserFromAD(dnOfUser, client)
	if err != nil {
		log.Printf("[ERROR] Error while deleting user from AD: %s", err)
		return fmt.Errorf("Error while deleting user from AD %s", err)
	}
	log.Printf("[DEBUG] User deleted from AD successfully: %s", name)
	return nil
}

func resourceADUserRead(d *schema.ResourceData, meta interface{}) error {
	domain := d.Get("domain").(string)
	orgunit := d.Get("orgunit").(string)
	firstname := d.Get("firstname").(string)
	lastname := d.Get("lastname").(string)
	name := fmt.Sprintf("%s %s", firstname, lastname)

	var dnOfUser string
	if orgunit != "" {
		dnOfUser += "ou=Users,ou=" + orgunit
	} else {
		dnOfUser += "cn=Users"
	}
	domainArr := strings.Split(domain, ".")
	for _, item := range domainArr {
		dnOfUser += ",dc=" + item
	}

	log.Printf("[DEBUG] Name of the DN is : %s", dnOfUser)
	log.Printf("[DEBUG] Searching the user in the AD : %s", name)

	client := meta.(*ldap.Conn)

	searchRequest := ldap.NewSearchRequest(
		dnOfUser, // The base dn to search
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(&(objectClass=User)(cn="+name+"))",                                     // The filter to apply
		[]string{"dn", "cn", "description", "givenName", "sn", "sAMAccountName"}, // A list attributes to retrieve
		nil,
	)

	sr, err := client.Search(searchRequest)
	if err != nil {
		log.Printf("[ERROR] Error while searching a user: %s", err)
		return fmt.Errorf("Error while searching a user: %s", err)
	}
	if len(sr.Entries) == 0 {
		log.Println("[ERROR] User was not found")
		d.SetId("")
	} else {
		if len(sr.Entries) > 1 {
			log.Printf("[ERROR] Error found ambigious values for user: %s", name)
			return fmt.Errorf("Error found ambigious values for user: %s", name)
		}
		user := sr.Entries[0]
		d.SetId(user.DN)
		d.Set("username", user.GetAttributeValue("sAMAccountName"))
		d.Set("name", user.GetAttributeValue("cn"))
		d.Set("firstname", user.GetAttributeValue("givenName"))
		d.Set("lastname", user.GetAttributeValue("sn"))
		d.Set("description", user.GetAttributeValue("description"))
	}
	return nil
}
