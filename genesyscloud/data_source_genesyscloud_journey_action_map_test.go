package genesyscloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/testrunner"
)

func TestAccDataJourneyActionMap(t *testing.T) {
	runDataJourneyActionMapTestCase(t, "find_by_name")
}

func runDataJourneyActionMapTestCase(t *testing.T, testCaseName string) {
	const resourceName = "genesyscloud_journey_action_map"
	testObjectName := testrunner.TestObjectIdPrefix + testCaseName
	testObjectFullName := resourceName + "." + testObjectName
	setupJourneyActionMap(t, testCaseName)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: ProviderFactories,
		Steps: testrunner.GenerateDataSourceTestSteps(resourceName, testCaseName, []resource.TestCheckFunc{
			resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttrPair("data."+testObjectFullName, "id", testObjectFullName, "id"),
				resource.TestCheckResourceAttr(testObjectFullName, "display_name", testObjectName+"_to_find"),
			),
		}),
	})
}
