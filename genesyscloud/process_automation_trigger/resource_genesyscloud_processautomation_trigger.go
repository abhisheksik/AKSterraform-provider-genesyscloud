package process_automation_trigger

import (
	"context"

	"fmt"
	"log"

	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/mypurecloud/platform-client-sdk-go/v105/platformclientv2"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// Registering our resource provider for export
func init() {
	gcloud.RegisterResource("genesyscloud_processautomation_trigger", resourceProcessAutomationTrigger())
	gcloud.RegisterDataSource("genesyscloud_processautomation_trigger", dataSourceProcessAutomationTrigger())
	gcloud.RegisterExporter("genesyscloud_processautomation_trigger", processAutomationTriggerExporter())
}

var (
	target = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"type": {
				Description: "Type of the target the trigger is configured to hit",
				Type:        schema.TypeString,
				Required:    true,
				ValidateFunc: validation.StringInSlice([]string{
					"Workflow",
				}, false),
			},
			"id": {
				Description: "Id of the target the trigger is configured to hit",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
)

/*
NOTE:
This resource currently does not use the Go SDk and instead makes API calls directly.
The Go SDK can not properly handle process automation triggers due the value and values
attributes in the matchCriteria object being listed as JsonNode in the swagger docs.
A JsonNode is a placeholder type with no nested values which creates problems in Go
because it can't properly determine a type for the value/values field.
*/
func resourceProcessAutomationTrigger() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud Process Automation Trigger`,

		CreateContext: gcloud.CreateWithPooledClient(createProcessAutomationTrigger),
		ReadContext:   gcloud.ReadWithPooledClient(readProcessAutomationTrigger),
		UpdateContext: gcloud.UpdateWithPooledClient(updateProcessAutomationTrigger),
		DeleteContext: gcloud.DeleteWithPooledClient(removeProcessAutomationTrigger),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description:  "Name of the Trigger",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 256),
			},
			"topic_name": {
				Description:  "Topic name that will fire trigger. Changing the topic_name attribute will cause the processautomation_trigger object to be dropped and recreated with a new ID. ",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 256),
			},
			"enabled": {
				Description: "Whether or not the trigger should be fired on events",
				Type:        schema.TypeBool,
				Required:    true,
			},
			"target": {
				Description: "Target the trigger will invoke when fired",
				Type:        schema.TypeSet,
				Optional:    false,
				Required:    true,
				MaxItems:    1,
				Elem:        target,
			},
			"match_criteria": {
				Description: "Match criteria that controls when the trigger will fire. NOTE: The match_criteria field type has changed from a complex object to a string. This was done to allow for complex JSON object definitions.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"event_ttl_seconds": {
				Description:  "How old an event can be to fire the trigger. Must be an number greater than or equal to 10. Only one of event_ttl_seconds or delay_by_seconds can be set.",
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(10),
			},
			"delay_by_seconds": {
				Description:  "How long to delay processing of a trigger after an event passes the match criteria. Must be an number between 60 and 900 inclusive. Only one of event_ttl_seconds or delay_by_seconds can be set.",
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(60, 900),
			},
			"description": {
				Description:  "A description of the trigger",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 512),
			},
		},
	}
}

func processAutomationTriggerExporter() *gcloud.ResourceExporter {
	return &gcloud.ResourceExporter{
		GetResourcesFunc: gcloud.GetAllWithPooledClient(getAllProcessAutomationTriggersResourceMap),
		RefAttrs: map[string]*gcloud.RefAttrSettings{
			"target.id": {RefType: "genesyscloud_flow"},
		},
	}
}

func createProcessAutomationTrigger(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	topic_name := d.Get("topic_name").(string)
	enabled := d.Get("enabled").(bool)
	eventTTLSeconds := d.Get("event_ttl_seconds").(int)
	delayBySeconds := d.Get("delay_by_seconds").(int)
	description := d.Get("description").(string)
	matchingCriteria := d.Get("match_criteria").(string)

	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	integAPI := platformclientv2.NewIntegrationsApiWithConfig(sdkConfig)

	if eventTTLSeconds > 0 && delayBySeconds > 0 {
		return diag.Errorf("Only one of event_ttl_seconds or delay_by_seconds can be set.")
	}

	log.Printf("Creating process automation trigger %s", name)

	triggerInput := &ProcessAutomationTrigger{
		TopicName:     &topic_name,
		Name:          &name,
		Target:        buildTarget(d),
		MatchCriteria: &matchingCriteria,
		Enabled:       &enabled,
		Description:   &description,
	}

	if eventTTLSeconds > 0 {
		triggerInput.EventTTLSeconds = &eventTTLSeconds
	}

	if delayBySeconds > 0 {
		triggerInput.DelayBySeconds = &delayBySeconds
	}

	diagErr := gcloud.RetryWhen(gcloud.IsStatus400, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		trigger, resp, err := postProcessAutomationTrigger(triggerInput, integAPI)

		if err != nil {
			return resp, diag.Errorf("Failed to create process automation trigger %s: %s", name, err)
		}
		d.SetId(*trigger.Id)

		log.Printf("Created process automation trigger %s %s", name, *trigger.Id)
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	return readProcessAutomationTrigger(ctx, d, meta)
}

func readProcessAutomationTrigger(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	integAPI := platformclientv2.NewIntegrationsApiWithConfig(sdkConfig)

	log.Printf("Reading process automation trigger %s", d.Id())

	return gcloud.WithRetriesForRead(ctx, d, func() *resource.RetryError {
		trigger, resp, getErr := getProcessAutomationTrigger(d.Id(), integAPI)
		if getErr != nil {
			if gcloud.IsStatus404(resp) {
				return resource.RetryableError(fmt.Errorf("Failed to read process automation trigger %s: %s", d.Id(), getErr))
			}
			return resource.NonRetryableError(fmt.Errorf("Failed to process read automation trigger %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, resourceProcessAutomationTrigger())

		if trigger.Name != nil {
			d.Set("name", *trigger.Name)
		} else {
			d.Set("name", nil)
		}

		if trigger.TopicName != nil {
			d.Set("topic_name", *trigger.TopicName)
		} else {
			d.Set("topic_name", nil)
		}

		d.Set("match_criteria", trigger.MatchCriteria)
		d.Set("target", flattenTarget(trigger.Target))

		if trigger.Enabled != nil {
			d.Set("enabled", *trigger.Enabled)
		} else {
			d.Set("enabled", nil)
		}

		if trigger.EventTTLSeconds != nil {
			d.Set("event_ttl_seconds", *trigger.EventTTLSeconds)
		} else {
			d.Set("event_ttl_seconds", nil)
		}

		if trigger.DelayBySeconds != nil {
			d.Set("delay_by_seconds", *trigger.DelayBySeconds)
		} else {
			d.Set("delay_by_seconds", nil)
		}

		if trigger.Description != nil {
			d.Set("description", *trigger.Description)
		} else {
			d.Set("description", nil)
		}

		log.Printf("Read process automation trigger %s %s", d.Id(), *trigger.Name)
		return cc.CheckState()
	})
}

func updateProcessAutomationTrigger(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	enabled := d.Get("enabled").(bool)
	eventTTLSeconds := d.Get("event_ttl_seconds").(int)
	delayBySeconds := d.Get("delay_by_seconds").(int)
	description := d.Get("description").(string)
	matchingCriteria := d.Get("match_criteria").(string)

	topic_name := d.Get("topic_name").(string)

	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	integAPI := platformclientv2.NewIntegrationsApiWithConfig(sdkConfig)

	log.Printf("Updating process automation trigger %s", name)

	diagErr := gcloud.RetryWhen(gcloud.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get the latest trigger version to send with PATCH
		trigger, resp, getErr := getProcessAutomationTrigger(d.Id(), integAPI)
		if getErr != nil {
			return resp, diag.Errorf("Failed to read process automation trigger %s: %s", d.Id(), getErr)
		}

		if eventTTLSeconds > 0 && delayBySeconds > 0 {
			return resp, diag.Errorf("Only one of event_ttl_seconds or delay_by_seconds can be set.")
		}

		triggerInput := &ProcessAutomationTrigger{
			TopicName:     &topic_name,
			Name:          &name,
			Enabled:       &enabled,
			Target:        buildTarget(d),
			MatchCriteria: &matchingCriteria,
			Version:       trigger.Version,
			Description:   &description,
		}

		if eventTTLSeconds > 0 {
			triggerInput.EventTTLSeconds = &eventTTLSeconds
		}

		if delayBySeconds > 0 {
			triggerInput.DelayBySeconds = &delayBySeconds
		}

		_, putResp, err := putProcessAutomationTrigger(d.Id(), triggerInput, integAPI)

		if err != nil {
			return putResp, diag.Errorf("Failed to update process automation trigger %s: %s", name, err)
		}
		return putResp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updated process automation trigger %s", name)
	return readProcessAutomationTrigger(ctx, d, meta)
}

func removeProcessAutomationTrigger(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	integAPI := platformclientv2.NewIntegrationsApiWithConfig(sdkConfig)

	log.Printf("Deleting process automation trigger %s", name)

	return gcloud.WithRetries(ctx, 30*time.Second, func() *resource.RetryError {
		resp, err := deleteProcessAutomationTrigger(d.Id(), integAPI)

		if err != nil {
			if gcloud.IsStatus404(resp) {
				log.Printf("process automation trigger already deleted %s", d.Id())
				return nil
			}
			return resource.RetryableError(fmt.Errorf("process automation trigger %s still exists", d.Id()))
		}
		return nil
	})
}

func buildTarget(d *schema.ResourceData) *Target {
	if target := d.Get("target"); target != nil {
		if targetList := target.(*schema.Set).List(); len(targetList) > 0 {
			targetMap := targetList[0].(map[string]interface{})

			targetType := targetMap["type"].(string)
			id := targetMap["id"].(string)

			return &Target{
				Type: &targetType,
				Id:   &id,
			}
		}
	}

	return &Target{}
}

func flattenTarget(inputTarget *Target) *schema.Set {
	if inputTarget == nil {
		return nil
	}

	targetSet := schema.NewSet(schema.HashResource(target), []interface{}{})

	flattendedTarget := make(map[string]interface{})
	flattendedTarget["id"] = *inputTarget.Id
	flattendedTarget["type"] = *inputTarget.Type
	targetSet.Add(flattendedTarget)

	return targetSet
}
