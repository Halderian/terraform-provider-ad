package ad

import ldap "gopkg.in/ldap.v3"

func addOrgUnitToAD(orgUnitName string, dnName string, adConn *ldap.Conn, desc string) error {
	addRequest := ldap.NewAddRequest(dnName, nil)
	addRequest.Attribute("objectClass", []string{"organizationalunit"})
	addRequest.Attribute("ou", []string{orgUnitName})
	if desc != "" {
		addRequest.Attribute("description", []string{desc})
	}
	err := adConn.Add(addRequest)
	if err != nil {
		return err
	}
	return nil
}

func deleteOrgUnitFromAD(dnName string, adConn *ldap.Conn) error {
	delRequest := ldap.NewDelRequest(dnName, nil)
	err := adConn.Del(delRequest)
	if err != nil {
		return err
	}
	return nil
}
