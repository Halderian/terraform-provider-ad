package ad

import ldap "gopkg.in/ldap.v2"

func addUserToAD(UserName string, password string, name string, dnName string, adConn *ldap.Conn, desc string) error {
	addRequest := ldap.NewAddRequest(dnName)
	addRequest.Attribute("objectClass", []string{"user"})
	addRequest.Attribute("sAMAccountName", []string{UserName})
	addRequest.Attribute("displayName", []string{name})
	addRequest.Attribute("name", []string{name})
	addRequest.Attribute("userPassword", []string{password})
	addRequest.Attribute("userAccountControl", []string{"512"})
	addRequest.Attribute("primaryGroupID", []string{"513"})
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
