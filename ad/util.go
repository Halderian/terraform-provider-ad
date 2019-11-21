package ad

import (
	"fmt"
	"log"
	"reflect"
	"regexp"
)

// parses a given distinguised name (DN) and returns the object GUID plus rest
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
	} else {
		log.Printf("[ERROR] No regex matched input: %s ", dn)
		return "", dn
	}
}

// generates an GUID query string by taking the hex string of an Object ID
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

// extracts the name part of a DN given the resource specific identifier (ou | cn)
func parseDN(dn string, identifier string) (string, string) {
	log.Printf("[DEBUG] Given DN string: %s ", dn)
	regex1 := regexp.MustCompile(fmt.Sprintf(`^(?i)%s=(?P<NAME>[\w- ]*),(?P<PARENT>.*)$`, identifier))

	if regex1.MatchString(dn) {
		res := regex1.FindStringSubmatch(dn)
		log.Printf("[DEBUG] Result of regex: %s ", res)
		return res[1], res[2]
	}

	return "", dn
}

// extracts the domain part of a DN
func extractDomainFromDN(dn string) string {
	log.Printf("[DEBUG] Given DN string: %s ", dn)
	regex1 := regexp.MustCompile(`(?i)dc=(\w*),?`)

	if regex1.MatchString(dn) {
		res := regex1.FindAllStringSubmatch(dn, -1)
		log.Printf("[DEBUG] Result of regex: %s ", res)
		result := ""
		for _, match := range res {
			result += fmt.Sprintf("dc=%s,", match[1])
		}
		result = result[:len(result)-1]
		log.Printf("[DEBUG] Result of operation: %s ", result)
		return result
	}

	return dn
}

func itemExists(arrayType interface{}, item interface{}) bool {
	arr := reflect.ValueOf(arrayType)

	if arr.Kind() != reflect.Array && arr.Kind() != reflect.Slice {
		log.Printf("[ERROR] Invalid data type %s", arr.Kind())
		panic("Invalid data-type")
	}

	for i := 0; i < arr.Len(); i++ {
		if arr.Index(i).Interface() == item {
			return true
		}
	}

	return false
}
