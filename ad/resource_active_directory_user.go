package ad

import (
	"fmt"
	"log"

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
				ForceNew:    false,
			},
			"password": {
				Type:        schema.TypeString,
				Description: "The login password of the user",
				Required:    true,
				Sensitive:   true,
				ForceNew:    false,
			},
			"parent": {
				Type:        schema.TypeString,
				Description: "The parent the domain belongs to. Could be either the DN of an OU or a DC.",
				Required:    true,
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
			"dn": {
				Type:        schema.TypeString,
				Description: "The distinguished name of the user",
				Computed:    true,
			},
			"groups": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},
		},
	}
}

func resourceADUserCreate(d *schema.ResourceData, meta interface{}) error {
	username := d.Get("username").(string)
	password := d.Get("password").(string)
	parent := d.Get("parent").(string)
	description := d.Get("description").(string)
	firstname := d.Get("firstname").(string)
	lastname := d.Get("lastname").(string)
	name := fmt.Sprintf("%s %s", firstname, lastname)

	dnOfUser := fmt.Sprintf("cn=%s,%s", name, parent)

	log.Printf("[DEBUG] Name of the DN is : %s", dnOfUser)
	log.Printf("[DEBUG] Adding the user to the AD : %s", name)

	client := meta.(*ldap.Conn)

	err := addUserToAD(username, firstname, lastname, dnOfUser, client, description)
	if err != nil {
		log.Printf("[ERROR] Error while adding a user to the AD : %s", err)
		return fmt.Errorf("Error while adding a user to the AD %s", err)
	}
	d.Set("dn", dnOfUser)
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
	parent := d.Get("parent").(string)
	firstname := d.Get("firstname").(string)
	lastname := d.Get("lastname").(string)
	name := fmt.Sprintf("%s %s", firstname, lastname)

	dnOfUser := fmt.Sprintf("cn=%s,%s", name, parent)

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
	var username string
	var searchParam string
	var parent string

	dnOfUser := d.Get("dn").(string)

	if dnOfUser == "" {
		username = d.Get("username").(string)
		parent = d.Get("parent").(string)

		dnOfUser += parent
		searchParam = "(sAMAccountName=" + username + ")"
	} else {
		searchParam = "(distinguishedName=" + dnOfUser + ")"
	}

	_, searchBaseDN := parseDN(dnOfUser, "cn")

	log.Printf("[DEBUG] Name of the DN is : %s", dnOfUser)
	log.Printf("[DEBUG] Searching the user in the AD : %s", username)

	client := meta.(*ldap.Conn)

	if d.Id() != "" {
		searchParam = "(objectGUID=" + generateObjectIdQueryString(d.Id()) + ")"
		searchBaseDN = extractDomainFromDN(dnOfUser)
	}

	log.Printf("[DEBUG] Search Parameters for user: %s ", searchParam)

	searchRequest := ldap.NewSearchRequest(
		searchBaseDN, // The base dn to search
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(&(objectClass=User)"+searchParam+")",                                               // The filter to apply
		[]string{"dn", "cn", "description", "givenName", "sn", "sAMAccountName", "memberOf"}, // A list attributes to retrieve
		nil,
	)

	searchRequest.Controls = append(searchRequest.Controls, &ldapControlServerExtendDN{})

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
			log.Printf("[ERROR] Error found ambigious values for user: %s", username)
			return fmt.Errorf("Error found ambigious values for user: %s", username)
		}
		var name string
		user := sr.Entries[0]
		userID, userDN := parseExtendedDN(user.DN)
		name, parent = parseDN(userDN, "cn")

		var userGroups []string
		for _, group := range user.GetAttributeValues("memberOf") {
			_, dn := parseExtendedDN(group)
			userGroups = append(userGroups, dn)
		}

		d.SetId(userID)
		d.Set("dn", userDN)
		d.Set("username", user.GetAttributeValue("sAMAccountName"))
		d.Set("name", name)
		d.Set("firstname", user.GetAttributeValue("givenName"))
		d.Set("lastname", user.GetAttributeValue("sn"))
		d.Set("description", user.GetAttributeValue("description"))
		d.Set("parent", parent)
		d.Set("groups", userGroups)
	}
	return nil
}
