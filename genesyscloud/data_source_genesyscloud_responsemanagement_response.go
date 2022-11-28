package genesyscloud

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v80/platformclientv2"
	"time"
)

func dataSourceResponsemanagementResponse() *schema.Resource {
	return &schema.Resource{
		Description: `Data source for Genesys Cloud Responsemanagement Response. Select a Responsemanagement Response by name.`,

		ReadContext: readWithPooledClient(dataSourceResponsemanagementResponseRead),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: `Responsemanagement Response name.`,
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
}

func dataSourceResponsemanagementResponseRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	responseManagementApi := platformclientv2.NewResponseManagementApiWithConfig(sdkConfig)

	name := d.Get("name").(string)

	return withRetries(ctx, 15*time.Second, func() *resource.RetryError {
		for pageNum := 1; ; pageNum++ {
			const pageSize = 100
			sdkresponseentitylisting, _, getErr := responseManagementApi.GetResponsemanagementResponses("", pageNum, pageSize, "")
			if getErr != nil {
				return resource.NonRetryableError(fmt.Errorf("Error requesting Responsemanagement Response %s: %s", name, getErr))
			}

			if sdkresponseentitylisting.Entities == nil || len(*sdkresponseentitylisting.Entities) == 0 {
				return resource.RetryableError(fmt.Errorf("No Responsemanagement Response found with name %s", name))
			}

			for _, entity := range *sdkresponseentitylisting.Entities {
				if entity.Name != nil && *entity.Name == name {
					d.SetId(*entity.Id)
					return nil
				}
			}
		}
	})
}
