package ad

import (
	"fmt"
	"log"
	"strings"

	ldap "gopkg.in/ldap.v3"

	"github.com/hashicorp/terraform/helper/schema"
)

func dataActiveDirectoryDomain() *schema.Resource {
	return &schema.Resource{
		Read: resourceADDomainRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the domain",
				Required:    true,
			},
			"parent": {
				Type:        schema.TypeString,
				Description: "The parent of the domain",
				Required:    true,
			},
			"dn": {
				Type:        schema.TypeString,
				Description: "The distinguished name of the domain",
				Computed:    true,
			},
		},
	}
}

func resourceADDomainRead(d *schema.ResourceData, meta interface{}) error {
	domainName := d.Get("name").(string)
	parent := d.Get("parent").(string)

	dnOfDomain := "dc=" + domainName
	if parent != "" {
		domainArr := strings.Split(parent, ".")
		for _, item := range domainArr {
			dnOfDomain += ",dc=" + item
		}
	}

	log.Printf("[DEBUG] Name of the DN is : %s ", dnOfDomain)
	log.Printf("[DEBUG] Searching the domain from the AD : %s ", domainName)

	client := meta.(*ldap.Conn)

	searchParam := "(distinguishedName=" + dnOfDomain + ")"

	if d.Id() != "" {
		searchParam = "(objectGUID=" + generateObjectIdQueryString(d.Id()) + ")"
	}

	log.Printf("[DEBUG] Search Parameters for domain: %s ", searchParam)

	searchRequest := ldap.NewSearchRequest(
		dnOfDomain, // The base dn to search
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(&(objectClass=domain)"+searchParam+")", // The filter to apply
		[]string{"dn", "dc"},                     // A list attributes to retrieve
		nil,
	)

	searchRequest.Controls = append(searchRequest.Controls, &ldapControlServerExtendDN{})

	sr, err := client.Search(searchRequest)
	if err != nil {
		log.Printf("[ERROR] Error while searching a domain: %s", err)
		return fmt.Errorf("Error while searching a domain: %s", err)
	}
	if len(sr.Entries) == 0 {
		log.Println("[ERROR] Domain was not found")
		d.SetId("")
	} else {
		if len(sr.Entries) > 1 {
			log.Printf("[ERROR] Error found ambigious values for domain: %s", domainName)
			return fmt.Errorf("Error found ambigious values for domain: %s", domainName)
		}
		domainRecord := sr.Entries[0]
		domainID, domainDN := parseExtendedDN(domainRecord.DN)
		d.SetId(domainID)
		d.Set("dn", domainDN)
		d.Set("name", domainRecord.GetAttributeValue("dc"))
		d.Set("parent", extractDomainFromDN(domainDN))
	}
	return nil
}
