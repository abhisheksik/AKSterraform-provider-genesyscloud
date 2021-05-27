package genesyscloud

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/platformclientv2"
)

func getAllCredentials(ctx context.Context, clientConfig *platformclientv2.Configuration) (ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(ResourceIDMetaMap)
	integrationAPI := platformclientv2.NewIntegrationsApiWithConfig(clientConfig)

	for pageNum := 1; ; pageNum++ {
		credentials, _, err := integrationAPI.GetIntegrationsCredentials(pageNum, 100)
		if err != nil {
			return nil, diag.Errorf("Failed to get page of credentials: %v", err)
		}

		if credentials.Entities == nil || len(*credentials.Entities) == 0 {
			break
		}

		for _, cred := range *credentials.Entities {
			resources[*cred.Id] = &ResourceMeta{Name: *cred.Name}
		}
	}

	return resources, nil
}

func credentialExporter() *ResourceExporter {
	return &ResourceExporter{
		GetResourcesFunc: getAllWithPooledClient(getAllCredentials),
		RefAttrs:         map[string]*RefAttrSettings{},
	}
}

func resourceCredential() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Credential",

		CreateContext: createWithPooledClient(createCredential),
		ReadContext:   readWithPooledClient(readCredential),
		UpdateContext: updateWithPooledClient(updateCredential),
		DeleteContext: deleteWithPooledClient(deleteCredential),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Credential name.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"credential_type_name": {
				Description: "Credential type name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func createCredential(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	name := d.Get("name").(string)
	cred_type := d.Get("credential_type_name").(string)

	sdkConfig := meta.(*providerMeta).ClientConfig
	integrationAPI := platformclientv2.NewIntegrationsApiWithConfig(sdkConfig)

	createCredential := platformclientv2.Credential{
		Name: &name,
		VarType: &platformclientv2.Credentialtype{
			Name: &cred_type,
		},
	}

	credential, _, err := integrationAPI.PostIntegrationsCredentials(createCredential)

	if err != nil {
		return diag.Errorf("Failed to create credential %s : %s", name, err)
	}

	d.SetId(*credential.Id)

	log.Printf("Created credential %s, %s", name, *credential.Id)
	return readCredential(ctx, d, meta)
}

func readCredential(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	integrationAPI := platformclientv2.NewIntegrationsApiWithConfig(sdkConfig)

	log.Printf("Reading credential %s", d.Id())
	currentCredential, resp, getErr := integrationAPI.GetIntegrationsCredential(d.Id())

	if getErr != nil {
		if isStatus404(resp) {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Failed to read credential %s: %s", d.Id(), getErr)
	}

	d.Set("name", *currentCredential.Name)
	d.Set("credential_type_name", *currentCredential.VarType.Name)

	log.Printf("Read credential %s %s", d.Id(), *currentCredential.Name)

	return nil
}

func updateCredential(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	cred_type := d.Get("credential_type_name").(string)

	sdkConfig := meta.(*providerMeta).ClientConfig
	integrationAPI := platformclientv2.NewIntegrationsApiWithConfig(sdkConfig)

	if d.HasChanges("name", "credential_type_name") {

		log.Printf("Updating credential %s", name)

		_, _, putErr := integrationAPI.PutIntegrationsCredential(d.Id(), platformclientv2.Credential{
			Name: &name,
			VarType: &platformclientv2.Credentialtype{
				Name: &cred_type,
			},
		})
		if putErr != nil {
			return diag.Errorf("Failed to update credential %s: %s", name, putErr)
		}
	}

	log.Printf("Updated credential %s %s", name, d.Id())
	return readCredential(ctx, d, meta)
}

func deleteCredential(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	integrationAPI := platformclientv2.NewIntegrationsApiWithConfig(sdkConfig)

	_, err := integrationAPI.DeleteIntegrationsCredential(d.Id())
	if err != nil {
		return diag.Errorf("Failed to delete the credential %s: %s", d.Id(), err)
	}
	log.Printf("Deleted credential %s", d.Id())
	return nil
}
