package ad

import (
	"fmt"
	"golang.org/x/text/encoding/unicode"
	ldap "gopkg.in/ldap.v2"
)

func addUserToAD(UserName string, firstname string, lastname string, dnName string, adConn *ldap.Conn, desc string) error {
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
	if desc != "" {
		addRequest.Attribute("description", []string{desc})
	}
	err := adConn.Add(addRequest)
	if err != nil {
		return err
	}
	return nil
}

func setUserPassword(dnName string, password string, adConn *ldap.Conn) error {
	utf16 := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM)
	// The password needs to be enclosed in quotes
	pwdEncoded, err := utf16.NewEncoder().String(fmt.Sprintf("\"%s\"", password))
	if err != nil {
		return err
	}

	passwordModifyRequest := &ldap.ModifyRequest{
		DN: dnName, // DN for the user we're resetting
		ReplaceAttributes: []ldap.PartialAttribute{
			{"unicodePwd", []string{pwdEncoded}},
		},
	}
	err = adConn.Modify(passwordModifyRequest)
	if err != nil {
		return err
	}
	return nil
}

func activateUser(dnName string, adConn *ldap.Conn) error {
	activateUserRequest := &ldap.ModifyRequest{
		DN: dnName, // DN for the user we're resetting
		ReplaceAttributes: []ldap.PartialAttribute{
			{"userAccountControl", []string{"512"}},
		},
	}
	err := adConn.Modify(activateUserRequest)
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
