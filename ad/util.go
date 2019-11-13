package ad

import (
	"log"
	"regexp"
)

func parseExtendedDN(dn string) (string, string) {
	log.Printf("[DEBUG] Given DN string: %s ", dn)
	regex1 := regexp.MustCompile(`^<GUID=(?P<GUID>.*)>;<SID=(?P<SID>.*)>;(?P<DN>.*)$`)
	regex2 := regexp.MustCompile(`^<GUID=(?P<GUID>.*)>;(?P<DN>.*)$`)

	if regex1.MatchString(dn) {
		res := regex1.FindStringSubmatch(dn)
		log.Printf("[DEBUG] Result of regex: %s ", res)
		return res[1], res[3]
	} else if regex2.MatchString(dn) {
		res := regex2.FindStringSubmatch(dn)
		log.Printf("[DEBUG] Result of regex: %s ", res)
		return res[1], res[2]
	}

	log.Printf("[ERROR] No regex matched input: %s ", dn)
	return dn, dn
}

func generateObjectIdQueryString(oID string) string {
	log.Printf("[DEBUG] Given objectID %s ", oID)
	result := "\\"
	index := 0
	for _, runeValue := range oID {
		result += string(runeValue)
		index += 1
		if index%2 == 0 {
			result += "\\"
		}
	}
	result = result[:len(result)-1]
	log.Printf("[DEBUG] Result objectID %s ", result)
	return result
}
