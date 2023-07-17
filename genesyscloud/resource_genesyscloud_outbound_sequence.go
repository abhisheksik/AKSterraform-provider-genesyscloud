package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/platform-client-sdk-go/v105/platformclientv2"
)

func resourceOutboundSequence() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud outbound sequence`,

		CreateContext: CreateWithPooledClient(createOutboundSequence),
		ReadContext:   ReadWithPooledClient(readOutboundSequence),
		UpdateContext: UpdateWithPooledClient(updateOutboundSequence),
		DeleteContext: DeleteWithPooledClient(deleteOutboundSequence),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			`name`: {
				Description: `Name of outbound sequence`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`campaign_ids`: {
				Description: `The ordered list of Campaigns that this CampaignSequence will run.`,
				Required:    true,
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			`status`: {
				Description:  `The current status of the CampaignSequence. A CampaignSequence can be turned 'on' or 'off' (default). Changing from "on" to "off" will cause the current sequence to drop and be recreated with a new ID.`,
				Optional:     true,
				Computed:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{`on`, `off`}, false),
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return (old == `complete` && new == `on`)
				},
			},
			`repeat`: {
				Description: `Indicates if a sequence should repeat from the beginning after the last campaign completes. Default is false.`,
				Optional:    true,
				Type:        schema.TypeBool,
			},
		},
		CustomizeDiff: customdiff.ForceNewIfChange("status", func(ctx context.Context, old, new, meta any) bool {
			return new.(string) == "off" && (old.(string) == "on" || old.(string) == "complete")
		}),
	}
}

func getAllOutboundSequence(_ context.Context, clientConfig *platformclientv2.Configuration) (ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(ResourceIDMetaMap)
	outboundApi := platformclientv2.NewOutboundApiWithConfig(clientConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		sdkcampaignsequenceentitylisting, _, getErr := outboundApi.GetOutboundSequences(pageSize, pageNum, true, "", "", "", "")
		if getErr != nil {
			return nil, diag.Errorf("Error requesting page of Outbound Sequence: %s", getErr)
		}

		if sdkcampaignsequenceentitylisting.Entities == nil || len(*sdkcampaignsequenceentitylisting.Entities) == 0 {
			break
		}

		for _, entity := range *sdkcampaignsequenceentitylisting.Entities {
			if *entity.Status != "off" && *entity.Status != "on" {
				*entity.Status = "off"
			}
			resources[*entity.Id] = &ResourceMeta{Name: *entity.Name}
		}
	}

	return resources, nil
}

func outboundSequenceExporter() *ResourceExporter {
	return &ResourceExporter{
		GetResourcesFunc: GetAllWithPooledClient(getAllOutboundSequence),
		RefAttrs: map[string]*RefAttrSettings{
			`campaign_ids`: {
				RefType: "genesyscloud_outbound_campaign",
			},
		},
	}
}

func createOutboundSequence(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	status := d.Get("status").(string)
	repeat := d.Get("repeat").(bool)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	outboundApi := platformclientv2.NewOutboundApiWithConfig(sdkConfig)

	sdkcampaignsequence := platformclientv2.Campaignsequence{
		Campaigns: buildSdkDomainEntityRefArr(d, "campaign_ids"),
		Repeat:    &repeat,
	}

	if name != "" {
		sdkcampaignsequence.Name = &name
	}

	// All campaigns sequences have to be created in an "off" state to start out with
	defaultStatus := "off"
	sdkcampaignsequence.Status = &defaultStatus

	log.Printf("Creating Outbound Sequence %s", name)
	outboundSequence, _, err := outboundApi.PostOutboundSequences(sdkcampaignsequence)
	if err != nil {
		return diag.Errorf("Failed to create Outbound Sequence %s: %s", name, err)
	}

	d.SetId(*outboundSequence.Id)
	log.Printf("Created Outbound Sequence %s %s", name, *outboundSequence.Id)

	// Campaigns sequences can be enabled after creation
	if status == "on" {
		d.Set("status", status)
		diag := updateOutboundSequence(ctx, d, meta)
		if diag != nil {
			return diag
		}
	}

	return readOutboundSequence(ctx, d, meta)
}

func updateOutboundSequence(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	status := d.Get("status").(string)
	repeat := d.Get("repeat").(bool)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	outboundApi := platformclientv2.NewOutboundApiWithConfig(sdkConfig)

	sdkcampaignsequence := platformclientv2.Campaignsequence{
		Campaigns: buildSdkDomainEntityRefArr(d, "campaign_ids"),
		Repeat:    &repeat,
	}

	if name != "" {
		sdkcampaignsequence.Name = &name
	}
	if status != "" {
		sdkcampaignsequence.Status = &status
	}

	log.Printf("Updating Outbound Sequence %s", name)
	diagErr := RetryWhen(IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current Outbound Sequence version
		outboundSequence, resp, getErr := outboundApi.GetOutboundSequence(d.Id())
		if getErr != nil {
			return resp, diag.Errorf("Failed to read Outbound Sequence %s: %s", d.Id(), getErr)
		}
		sdkcampaignsequence.Version = outboundSequence.Version
		outboundSequence, _, updateErr := outboundApi.PutOutboundSequence(d.Id(), sdkcampaignsequence)
		if updateErr != nil {
			return resp, diag.Errorf("Failed to update Outbound Sequence %s: %s", name, updateErr)
		}
		return nil, nil
	})
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updated Outbound Sequence %s", name)
	return readOutboundSequence(ctx, d, meta)
}

func readOutboundSequence(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	outboundApi := platformclientv2.NewOutboundApiWithConfig(sdkConfig)

	log.Printf("Reading Outbound Sequence %s", d.Id())

	return WithRetriesForRead(ctx, d, func() *resource.RetryError {
		sdkcampaignsequence, resp, getErr := outboundApi.GetOutboundSequence(d.Id())
		if getErr != nil {
			if IsStatus404(resp) {
				return resource.RetryableError(fmt.Errorf("Failed to read Outbound Sequence %s: %s", d.Id(), getErr))
			}
			return resource.NonRetryableError(fmt.Errorf("Failed to read Outbound Sequence %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, resourceOutboundSequence())

		if sdkcampaignsequence.Name != nil {
			d.Set("name", *sdkcampaignsequence.Name)
		}
		if sdkcampaignsequence.Campaigns != nil {
			d.Set("campaign_ids", sdkDomainEntityRefArrToList(*sdkcampaignsequence.Campaigns))
		}
		if sdkcampaignsequence.Status != nil {
			d.Set("status", *sdkcampaignsequence.Status)
		}
		if sdkcampaignsequence.Repeat != nil {
			d.Set("repeat", *sdkcampaignsequence.Repeat)
		}

		log.Printf("Read Outbound Sequence %s %s", d.Id(), *sdkcampaignsequence.Name)

		return cc.CheckState()
	})
}

func deleteOutboundSequence(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	outboundApi := platformclientv2.NewOutboundApiWithConfig(sdkConfig)

	diagErr := RetryWhen(IsStatus400, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		log.Printf("Deleting Outbound Sequence")
		resp, err := outboundApi.DeleteOutboundSequence(d.Id())
		if err != nil {
			return resp, diag.Errorf("Failed to delete Outbound Sequence: %s", err)
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	return WithRetries(ctx, 30*time.Second, func() *resource.RetryError {
		_, resp, err := outboundApi.GetOutboundSequence(d.Id())
		if err != nil {
			if IsStatus404(resp) {
				// Outbound Sequence deleted
				log.Printf("Deleted Outbound Sequence %s", d.Id())
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("Error deleting Outbound Sequence %s: %s", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("Outbound Sequence %s still exists", d.Id()))
	})
}
