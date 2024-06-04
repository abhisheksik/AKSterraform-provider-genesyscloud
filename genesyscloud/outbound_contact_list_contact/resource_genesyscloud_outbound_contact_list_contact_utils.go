package outbound_contact_list_contact

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v130/platformclientv2"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
)

func buildWritableContactFromResourceData(d *schema.ResourceData) *platformclientv2.Writabledialercontact {
	contactListId := d.Get("contact_list_id").(string)
	contactId, _ := d.Get("id").(string)
	if contactId == "" {
		contactId = uuid.NewString() // TODO - determine if the api will compute this
	}
	callable := d.Get("callable").(bool)

	// TODO - add the rest of the fields
	var contactRequest = &platformclientv2.Writabledialercontact{
		Id:            &contactId,
		ContactListId: &contactListId,
		Callable:      &callable,
	}

	if dataMap, ok := d.Get("data").(map[string]string); ok {
		contactRequest.Data = &dataMap
	}

	contactRequest.PhoneNumberStatus = buildPhoneNumberStatus(d)

	contactRequest.ContactableStatus = buildContactableStatus(d)

	return contactRequest
}

func buildContactableStatus(d *schema.ResourceData) *map[string]platformclientv2.Contactablestatus {
	contactableStatus, ok := d.Get("contactable_status").(*schema.Set)
	if !ok {
		return nil
	}

	var contactableStatusMap map[string]platformclientv2.Contactablestatus

	contactableStatusList := contactableStatus.List()
	for _, status := range contactableStatusList {
		currentStatusMap := status.(map[string]any)
		mediaType := currentStatusMap["media_type"].(string)
		contactable := currentStatusMap["contactable"].(bool)
		var columnStatusMap map[string]platformclientv2.Columnstatus
		if columnStatus, ok := currentStatusMap["column_status"].(*schema.Set); ok {
			columnStatusList := columnStatus.List()
			for _, status := range columnStatusList {
				currentColumnStatusMap := status.(map[string]any)
				column := currentColumnStatusMap["column"].(string)
				columnContactable := currentColumnStatusMap["contactable"].(bool)
				columnStatusMap[column] = platformclientv2.Columnstatus{
					Contactable: &columnContactable,
				}
			}
		}
		contactableStatusMap[mediaType] = platformclientv2.Contactablestatus{
			Contactable:  &contactable,
			ColumnStatus: &columnStatusMap,
		}
	}

	return nil
}

func buildPhoneNumberStatus(d *schema.ResourceData) *map[string]platformclientv2.Phonenumberstatus {
	phoneNumberStatus, ok := d.Get("phone_number_status").(*schema.Set)
	if !ok {
		return nil
	}

	var phoneNumberStatusMap map[string]platformclientv2.Phonenumberstatus

	phoneNumberStatusList := phoneNumberStatus.List()
	for _, status := range phoneNumberStatusList {
		statusMap := status.(map[string]any)
		key := statusMap["key"].(string)
		callable, _ := statusMap["callable"].(bool)
		phoneNumberStatusMap[key] = platformclientv2.Phonenumberstatus{
			Callable: &callable,
		}
	}

	return &phoneNumberStatusMap
}

func flattenPhoneNumberStatus(phoneNumberStatus *map[string]platformclientv2.Phonenumberstatus) *schema.Set {
	pnsSet := schema.NewSet(schema.HashResource(phoneNumberStatusResource), []interface{}{})
	for k, v := range *phoneNumberStatus {
		pns := make(map[string]any)
		pns["key"] = k
		resourcedata.SetMapValueIfNotNil(pns, "callable", v.Callable)
		pnsSet.Add(pns)
	}
	return pnsSet
}

func flattenContactableStatus(contactableStatus *map[string]platformclientv2.Contactablestatus) *schema.Set {
	csSet := schema.NewSet(schema.HashResource(contactableStatusResource), []interface{}{})
	for k, v := range *contactableStatus {
		cs := make(map[string]any)
		cs["media_type"] = k
		cs["contactable"] = *v.Contactable
		cs["column_status"] = flattenColumnStatus(v.ColumnStatus)
		csSet.Add(cs)
	}
	return csSet
}

func flattenColumnStatus(columnStatus *map[string]platformclientv2.Columnstatus) *schema.Set {
	if columnStatus == nil {
		return nil
	}
	csSet := schema.NewSet(schema.HashResource(columnStatusResource), []interface{}{})
	for k, v := range *columnStatus {
		cs := make(map[string]any)
		cs["column"] = k
		cs["contactable"] = *v.Contactable
		csSet.Add(cs)
	}
	return csSet
}

const contactIdSplitCharacter = "_-_"

func createContactId(contactListId, contactId string) string {
	return contactListId + contactIdSplitCharacter + contactId
}

func parseContactListIdAndContactId(id string) (string, string, error) {
	ids := strings.Split(id, contactIdSplitCharacter)
	if len(ids) != 2 {
		return "", "", fmt.Errorf("expected to parse contact list id and contact id from string %s, splitting by '%s'. Got %v", id, contactIdSplitCharacter, len(ids))
	}
	return ids[0], ids[1], nil
}
