package ad

import (
	"fmt"
	ldap "gopkg.in/ldap.v2"
)

func addUserToAD(UserName string, password string, firstname string, lastname string, dnName string, adConn *ldap.Conn, desc string) error {
	userFullName := fmt.Sprintf("%s %s", firstname, lastname)
	addRequest := ldap.NewAddRequest(dnName)
	addRequest.Attribute("objectClass", []string{"user"})
	addRequest.Attribute("cn", []string{userFullName})
	addRequest.Attribute("displayName", []string{userFullName})
	addRequest.Attribute("givenName", []string{firstname})
	addRequest.Attribute("instanceType", []string{"4"})
	addRequest.Attribute("name", []string{userFullName})
	addRequest.Attribute("sAMAccountName", []string{UserName})
	addRequest.Attribute("sn", []string{lastname})
	addRequest.Attribute("userAccountControl", []string{"544"})
	addRequest.Attribute("userPassword", []string{password})
	if desc != "" {
		addRequest.Attribute("description", []string{desc})
	}
	err := adConn.Add(addRequest)
	if err != nil {
		return err
	}
	return nil
}

func deleteUserFromAD(dnName string, adConn *ldap.Conn) error {
	delRequest := ldap.NewDelRequest(dnName, nil)
	err := adConn.Del(delRequest)
	if err != nil {
		return err
	}
	return nil
}
