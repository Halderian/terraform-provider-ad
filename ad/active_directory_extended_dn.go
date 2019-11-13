package ad

import (
	"github.com/go-asn1-ber/asn1-ber"
)

// ldapControlServerPolicyHints mplements ldap.Control
type ldapControlServerExtendDN struct {
	Critical bool
	Flag     int
}

// GetControlType implements ldap.Control
func (c *ldapControlServerExtendDN) GetControlType() string {
	return "1.2.840.113556.1.4.529"
}

// Encode implements ldap.Control
func (c *ldapControlServerExtendDN) Encode() *ber.Packet {
	packet := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "Control")
	packet.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, c.GetControlType(), "Control Type (LDAP_SERVER_EXTENDED_DN_OID)"))
	packet.AppendChild(ber.NewBoolean(ber.ClassUniversal, ber.TypePrimitive, ber.TagBoolean, c.Critical, "Criticality"))

	p2 := ber.Encode(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, nil, "Control Value (Extend DN)")
	seq := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "ExtendDNRequestValue")
	seq.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagInteger, c.Flag, "Flags"))
	p2.AppendChild(seq)
	packet.AppendChild(p2)

	return packet
}

// String implements ldap.Control
func (c *ldapControlServerExtendDN) String() string {
	return "Extend DN request with objectGUID and objectSID: " + c.GetControlType()
}
