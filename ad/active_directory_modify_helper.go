package ad

import ldap "gopkg.in/ldap.v3"

func updateADEntry(entryDN string, attribute string, newValue string, adConn *ldap.Conn) error {
	updateRequest := ldap.NewModifyRequest(entryDN, nil)
	updateRequest.Replace(attribute, []string{newValue})
	err := adConn.Modify(updateRequest)
	if err != nil {
		return err
	}
	return nil
}

func renameADEntry(entryDN string, newName string, adConn *ldap.Conn) error {
	moveRequest := ldap.NewModifyDNRequest(entryDN, newName, true, "")
	err := adConn.ModifyDN(moveRequest)
	if err != nil {
		return err
	}
	return nil
}

func moveADEntry(entryDN string, entryName string, newParentDN string, adConn *ldap.Conn) error {
	moveRequest := ldap.NewModifyDNRequest(entryDN, entryName, true, newParentDN)
	err := adConn.ModifyDN(moveRequest)
	if err != nil {
		return err
	}
	return nil
}
