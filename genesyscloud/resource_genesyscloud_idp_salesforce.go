package genesyscloud

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/platformclientv2"
)

func getAllIdpSalesforce(ctx context.Context, clientConfig *platformclientv2.Configuration) (ResourceIDNameMap, diag.Diagnostics) {
	idpAPI := platformclientv2.NewIdentityProviderApiWithConfig(clientConfig)
	resources := make(map[string]string)

	_, resp, getErr := idpAPI.GetIdentityprovidersSalesforce()
	if getErr != nil {
		if resp != nil && resp.StatusCode == 404 {
			// Don't export if config doesn't exist
			return resources, nil
		}
		return nil, diag.Errorf("Failed to get IDP Salesforce: %v", getErr)
	}

	resources["0"] = "salesforce"
	return resources, nil
}

func idpSalesforceExporter() *ResourceExporter {
	return &ResourceExporter{
		GetResourcesFunc: getAllWithPooledClient(getAllIdpSalesforce),
		RefAttrs:         map[string]*RefAttrSettings{}, // No references
	}
}

func resourceIdpSalesforce() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Single Sign-on Salesforce Identity Provider. See this page for detailed configuration instructions: https://help.mypurecloud.com/articles/add-salesforce-as-a-single-sign-on-provider/",

		CreateContext: createWithPooledClient(createIdpSalesforce),
		ReadContext:   readWithPooledClient(readIdpSalesforce),
		UpdateContext: updateWithPooledClient(updateIdpSalesforce),
		DeleteContext: deleteWithPooledClient(deleteIdpSalesforce),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"certificates": {
				Description: "PEM or DER encoded public X.509 certificates for SAML signature validation.",
				Type:        schema.TypeSet,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"issuer_uri": {
				Description: "Issuer URI provided by Salesforce.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"target_uri": {
				Description: "Target URI provided by Salesforce.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"disabled": {
				Description: "True if Salesforce is disabled.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
		},
	}
}

func createIdpSalesforce(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("Creating IDP Salesforce")
	d.SetId("salesforce")
	return updateIdpSalesforce(ctx, d, meta)
}

func readIdpSalesforce(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	idpAPI := platformclientv2.NewIdentityProviderApiWithConfig(sdkConfig)

	log.Printf("Reading IDP Salesforce")
	salesforce, resp, getErr := idpAPI.GetIdentityprovidersSalesforce()
	if getErr != nil {
		if resp != nil && resp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Failed to read IDP Salesforce: %s", getErr)
	}

	if salesforce.Certificate != nil {
		d.Set("certificates", stringListToSet([]string{*salesforce.Certificate}))
	} else if salesforce.Certificates != nil {
		d.Set("certificates", stringListToSet(*salesforce.Certificates))
	} else {
		d.Set("certificates", nil)
	}

	if salesforce.IssuerURI != nil {
		d.Set("issuer_uri", *salesforce.IssuerURI)
	} else {
		d.Set("issuer_uri", nil)
	}

	if salesforce.SsoTargetURI != nil {
		d.Set("target_uri", *salesforce.SsoTargetURI)
	} else {
		d.Set("target_uri", nil)
	}

	if salesforce.Disabled != nil {
		d.Set("disabled", *salesforce.Disabled)
	} else {
		d.Set("disabled", nil)
	}

	log.Printf("Read IDP Salesforce")
	return nil
}

func updateIdpSalesforce(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	issuerUri := d.Get("issuer_uri").(string)
	targetUri := d.Get("target_uri").(string)
	disabled := d.Get("disabled").(bool)

	sdkConfig := meta.(*providerMeta).ClientConfig
	idpAPI := platformclientv2.NewIdentityProviderApiWithConfig(sdkConfig)

	log.Printf("Updating IDP Salesforce")
	update := platformclientv2.Salesforce{
		IssuerURI:    &issuerUri,
		SsoTargetURI: &targetUri,
		Disabled:     &disabled,
	}

	certificates := buildSdkStringList(d, "certificates")
	if certificates != nil {
		if len(*certificates) == 1 {
			update.Certificate = &(*certificates)[0]
		} else {
			update.Certificates = certificates
		}
	}

	_, _, err := idpAPI.PutIdentityprovidersSalesforce(update)
	if err != nil {
		return diag.Errorf("Failed to update IDP Salesforce: %s", err)
	}

	log.Printf("Updated IDP Salesforce")
	return readIdpSalesforce(ctx, d, meta)
}

func deleteIdpSalesforce(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	idpAPI := platformclientv2.NewIdentityProviderApiWithConfig(sdkConfig)

	log.Printf("Deleting IDP Salesforce")
	_, _, err := idpAPI.DeleteIdentityprovidersSalesforce()
	if err != nil {
		return diag.Errorf("Failed to delete IDP Salesforce: %s", err)
	}
	log.Printf("Deleted IDP Salesforce")
	return nil
}
