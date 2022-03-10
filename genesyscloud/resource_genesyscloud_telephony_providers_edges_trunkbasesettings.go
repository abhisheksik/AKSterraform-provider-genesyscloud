package genesyscloud

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/platform-client-sdk-go/v56/platformclientv2"
	"log"
	"net/http"
	"time"
)

func resourceTrunkBaseSettings() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Trunk Base Settings",

		CreateContext: createWithPooledClient(createTrunkBaseSettings),
		ReadContext:   readWithPooledClient(readTrunkBaseSettings),
		UpdateContext: updateWithPooledClient(updateTrunkBaseSettings),
		DeleteContext: deleteWithPooledClient(deleteTrunkBaseSettings),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name of the entity.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"state": {
				Description: "The resource's state.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"description": {
				Description: "The resource's description.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"trunk_meta_base_id": {
				Description: "The meta-base this trunk is based on.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"properties": {
				Description:      "trunk base settings properties",
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: suppressEquivalentJsonDiffs,
			},
			"trunk_type": {
				Description:  "The type of this trunk base.Valid values: EXTERNAL, PHONE, EDGE.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"EXTERNAL", "PHONE", "EDGE"}, false),
			},
			"managed": {
				Description: "Is this trunk being managed remotely. This property is synchronized with the managed property of the Edge Group to which it is assigned.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
		},
		CustomizeDiff: customizeTrunkBaseSettingsPropertiesDiff,
	}
}

func createTrunkBaseSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	trunkMetaBase := buildSdkDomainEntityRef(d, "trunk_meta_base_id")
	properties := buildBaseSettingsProperties(d)

	trunkType := d.Get("trunk_type").(string)
	managed := d.Get("managed").(bool)

	trunkBase := platformclientv2.Trunkbase{
		Name:          &name,
		TrunkMetabase: trunkMetaBase,
		TrunkType:     &trunkType,
		Managed:       &managed,
		Properties:    properties,
	}

	if description != "" {
		trunkBase.Description = &description
	}

	sdkConfig := meta.(*providerMeta).ClientConfig
	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	log.Printf("Creating trunk base settings %s", name)
	trunkBaseSettings, _, err := edgesAPI.PostTelephonyProvidersEdgesTrunkbasesettings(trunkBase)
	if err != nil {
		return diag.Errorf("Failed to create trunk base settings %s: %s", name, err)
	}

	d.SetId(*trunkBaseSettings.Id)

	log.Printf("Created trunk base settings %s", *trunkBaseSettings.Id)

	return readTrunkBaseSettings(ctx, d, meta)
}

func updateTrunkBaseSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	trunkMetaBase := buildSdkDomainEntityRef(d, "trunk_meta_base_id")
	properties := buildBaseSettingsProperties(d)
	trunkType := d.Get("trunk_type").(string)
	managed := d.Get("managed").(bool)
	id := d.Id()

	trunkBase := platformclientv2.Trunkbase{
		Id:            &id,
		Name:          &name,
		TrunkMetabase: trunkMetaBase,
		TrunkType:     &trunkType,
		Managed:       &managed,
		Properties:    properties,
	}

	if description != "" {
		trunkBase.Description = &description
	}

	sdkConfig := meta.(*providerMeta).ClientConfig
	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	diagErr := retryWhen(isVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get the latest version of the setting
		trunkBaseSettings, resp, getErr := edgesAPI.GetTelephonyProvidersEdgesTrunkbasesetting(d.Id(), true)
		if getErr != nil {
			if isStatus404(resp) {
				return resp, diag.Errorf("The trunk base settings does not exist %s: %s", d.Id(), getErr)
			}
			return resp, diag.Errorf("Failed to read trunk base settings %s: %s", d.Id(), getErr)
		}
		trunkBase.Version = trunkBaseSettings.Version

		log.Printf("Updating trunk base settings %s", name)
		trunkBaseSettings, resp, err := edgesAPI.PutTelephonyProvidersEdgesTrunkbasesetting(d.Id(), trunkBase)
		if err != nil {
			respString := ""
			if resp != nil {
				respString = resp.String()
			}
			return resp, diag.Errorf("Failed to update trunk base settings %s: %s %v", name, err, respString)
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	// Get the latest version of the setting
	trunkBaseSettings, resp, getErr := edgesAPI.GetTelephonyProvidersEdgesTrunkbasesetting(d.Id(), true)
	if getErr != nil {
		if isStatus404(resp) {
			return nil
		}
		return diag.Errorf("Failed to read trunk base settings %s: %s", d.Id(), getErr)
	}
	trunkBase.Version = trunkBaseSettings.Version

	log.Printf("Updating trunk base settings %s", name)
	trunkBaseSettings, resp, err := edgesAPI.PutTelephonyProvidersEdgesTrunkbasesetting(d.Id(), trunkBase)
	if err != nil {
		respString := ""
		if resp != nil {
			respString = resp.String()
		}
		return diag.Errorf("Failed to update trunk base settings %s: %s %v", name, err, respString)
	}

	log.Printf("Updated trunk base settings %s", *trunkBaseSettings.Id)

	return readTrunkBaseSettings(ctx, d, meta)
}

func readTrunkBaseSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	log.Printf("Reading trunk base settings %s", d.Id())
	return withRetriesForRead(ctx, 30*time.Second, d, func() *resource.RetryError {
		trunkBaseSettings, resp, getErr := edgesAPI.GetTelephonyProvidersEdgesTrunkbasesetting(d.Id(), true)
		if getErr != nil {
			if isStatus404(resp) {
				return resource.RetryableError(fmt.Errorf("Failed to read trunk base settings %s: %s", d.Id(), getErr))
			}
			return resource.NonRetryableError(fmt.Errorf("Failed to read trunk base settings %s: %s", d.Id(), getErr))
		}

		cc := NewConsistencyCheck(d)
		d.Set("name", *trunkBaseSettings.Name)
		d.Set("state", *trunkBaseSettings.State)
		if trunkBaseSettings.Description != nil {
			d.Set("description", *trunkBaseSettings.Description)
		}
		if trunkBaseSettings.Managed != nil {
			d.Set("managed", *trunkBaseSettings.Managed)
		}
		if trunkBaseSettings.TrunkMetabase != nil {
			d.Set("trunk_meta_base_id", *trunkBaseSettings.TrunkMetabase.Id)
		}
		d.Set("trunk_type", *trunkBaseSettings.TrunkType)

		d.Set("properties", nil)
		if trunkBaseSettings.Properties != nil {
			properties, err := flattenBaseSettingsProperties(trunkBaseSettings.Properties)
			if err != nil {
				return resource.NonRetryableError(fmt.Errorf("%v", err))
			}
			d.Set("properties", properties)
		}

		log.Printf("Read trunk base settings %s %s", d.Id(), *trunkBaseSettings.Name)

		return cc.CheckErr()
	})
}

func deleteTrunkBaseSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	diagErr := retryWhen(isStatus400, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		log.Printf("Deleting trunk base settings")
		resp, err := edgesAPI.DeleteTelephonyProvidersEdgesTrunkbasesetting(d.Id())
		if err != nil {
			return resp, diag.Errorf("Failed to delete trunk base settings: %s", err)
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	return withRetries(ctx, 30*time.Second, func() *resource.RetryError {
		trunkBaseSettings, resp, err := edgesAPI.GetTelephonyProvidersEdgesTrunkbasesetting(d.Id(), true)
		if err != nil {
			if isStatus404(resp) {
				// trunk base settings deleted
				log.Printf("Deleted trunk base settings %s", d.Id())
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("Error deleting trunk base settings %s: %s", d.Id(), err))
		}

		if trunkBaseSettings.State != nil && *trunkBaseSettings.State == "deleted" {
			// trunk base settings deleted
			log.Printf("Deleted trunk base settings %s", d.Id())
			return nil
		}

		return resource.RetryableError(fmt.Errorf("trunk base settings %s still exists", d.Id()))
	})
}

func getAllTrunkBaseSettings(ctx context.Context, sdkConfig *platformclientv2.Configuration) (ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(ResourceIDMetaMap)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		trunkBaseSettings, _, getErr := getTelephonyProvidersEdgesTrunkbasesettings(sdkConfig, pageNum, pageSize)
		if getErr != nil {
			return nil, diag.Errorf("Failed to get page of trunk base settings: %v", getErr)
		}

		if trunkBaseSettings.Entities == nil || len(*trunkBaseSettings.Entities) == 0 {
			break
		}

		for _, trunkBaseSetting := range *trunkBaseSettings.Entities {
			if trunkBaseSetting.State != nil && *trunkBaseSetting.State != "deleted" {
				resources[*trunkBaseSetting.Id] = &ResourceMeta{Name: *trunkBaseSetting.Name}
			}
		}
	}

	return resources, nil
}

// The SDK function is too cumbersome because of the various boolean query parameters.
// This function was written in order to leave them out and make a single API call
func getTelephonyProvidersEdgesTrunkbasesettings(sdkConfig *platformclientv2.Configuration, pageNumber int, pageSize int) (*platformclientv2.Trunkbaseentitylisting, *platformclientv2.APIResponse, error) {
	headerParams := make(map[string]string)
	if sdkConfig.AccessToken != ""{
		headerParams["Authorization"] =  "Bearer " + sdkConfig.AccessToken
	}
	// add default headers if any
	for key := range sdkConfig.DefaultHeader {
		headerParams[key] = sdkConfig.DefaultHeader[key]
	}

	queryParams := make(map[string]string)
	queryParams["pageNumber"] = sdkConfig.APIClient.ParameterToString(pageNumber, "")
	queryParams["pageSize"] = sdkConfig.APIClient.ParameterToString(pageSize, "")
	
	// to determine the Content-Type header
	httpContentTypes := []string{ "application/json" }

	// set Content-Type header
	httpContentType := sdkConfig.APIClient.SelectHeaderContentType(httpContentTypes)
	if httpContentType != "" {
		headerParams["Content-Type"] = httpContentType
	}

	// set Accept header
	httpHeaderAccept := sdkConfig.APIClient.SelectHeaderAccept([]string{
		"application/json",
	})
	if httpHeaderAccept != "" {
		headerParams["Accept"] = httpHeaderAccept
	}
	var successPayload *platformclientv2.Trunkbaseentitylisting
	path := sdkConfig.BasePath + "/api/v2/telephony/providers/edges/trunkbasesettings"
	response, err := sdkConfig.APIClient.CallAPI(path, http.MethodGet, nil, headerParams, queryParams, nil, "", nil)
	if err != nil {
		return nil, nil, err
	}

	if response.Error != nil {
		err = errors.New(response.ErrorMessage)
	} else {
		err = json.Unmarshal(response.RawBody, &successPayload)
	}
	return successPayload, response, err
}

func trunkBaseSettingsExporter() *ResourceExporter {
	return &ResourceExporter{
		GetResourcesFunc: getAllWithPooledClient(getAllTrunkBaseSettings),
		RefAttrs:         map[string]*RefAttrSettings{},
	}
}
