package ad

import ldap "gopkg.in/ldap.v3"

func addGroupToAD(groupName string, dnName string, typeOfGroup string, adConn *ldap.Conn, desc string) error {
	addRequest := ldap.NewAddRequest(dnName, nil)
	addRequest.Attribute("objectClass", []string{"group"})
	addRequest.Attribute("sAMAccountName", []string{groupName})
	if desc != "" {
		addRequest.Attribute("description", []string{desc})
	}
	if typeOfGroup == "LOCAL" {
		addRequest.Attribute("groupType", []string{"-2147483644"})
	} else {
		addRequest.Attribute("groupType", []string{"-2147483646"})
	}
	err := adConn.Add(addRequest)
	if err != nil {
		return err
	}
	return nil
}

func deleteGroupFromAD(groupDN string, adConn *ldap.Conn) error {
	delRequest := ldap.NewDelRequest(groupDN, nil)
	err := adConn.Del(delRequest)
	if err != nil {
		return err
	}
	return nil
}

func updateGroupDetails(groupDN string, attribute string, newValue string, adConn *ldap.Conn) error {
	updateRequest := ldap.NewModifyRequest(groupDN, nil)
	updateRequest.Replace(attribute, []string{newValue})
	err := adConn.Modify(updateRequest)
	if err != nil {
		return err
	}
	return nil
}

func addMemberToGroup(groupDN string, memberDN string, adConn *ldap.Conn) error {
	modifyRequest := ldap.NewModifyRequest(groupDN, nil)
	modifyRequest.Add("member", []string{memberDN})
	err := adConn.Modify(modifyRequest)
	if err != nil {
		return err
	}
	return nil
}

func removeMemberFromGroup(groupDN string, memberDN string, adConn *ldap.Conn) error {
	modifyRequest := ldap.NewModifyRequest(groupDN, nil)
	modifyRequest.Delete("member", []string{memberDN})
	err := adConn.Modify(modifyRequest)
	if err != nil {
		return err
	}
	return nil
}
