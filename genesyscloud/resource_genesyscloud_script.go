package genesyscloud

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v95/platformclientv2"
)

func resourceScript() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Script",

		CreateContext: createWithPooledClient(createScript),
		ReadContext:   readWithPooledClient(readScript),
		DeleteContext: deleteWithPooledClient(deleteScript),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"filename": {
				Description: "Path to the script file to upload.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"script_name": {
				Description: "Display name for the script.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
		},
	}
}

func createScript(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var (
		sdkConfig   = meta.(*ProviderMeta).ClientConfig
		scriptsAPI  = platformclientv2.NewScriptsApiWithConfig(sdkConfig)
		accessToken = scriptsAPI.Configuration.AccessToken

		bodyBuf = bytes.Buffer{}
		w       = multipart.NewWriter(&bodyBuf)
	)

	postUrl := scriptsAPI.Configuration.BasePath + "/uploads/v2/scripter"
	postUrl = strings.Replace(postUrl, "api", "apps", -1)

	fileName := d.Get("filename").(string)
	scriptName := d.Get("script_name").(string)

	duplicateName, err := scriptExistsWithName(scriptName, meta)
	if err != nil {
		return diag.Errorf("%v", err)
	}

	if duplicateName {
		return diag.Errorf("Script with name %s already exists. Please provide a unique name.", scriptName)
	}

	if err := createScriptFormData(fileName, scriptName, &bodyBuf, w); err != nil {
		return diag.Errorf("%v", err)
	}

	// using newrequest
	client := &http.Client{}
	r, _ := http.NewRequest(http.MethodPost, postUrl, &bodyBuf)
	r.Header.Set("Authorization", "Bearer "+accessToken)
	r.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := client.Do(r)
	if err != nil {
		return diag.Errorf("%v", err)
	}
	if resp.StatusCode != 200 {
		return diag.Errorf("error: %v", resp.Status)
	}

	// Need to get script ID back from POST request
	d.SetId(uuid.NewString())
	return readScript(ctx, d, meta)
}

func readScript(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	_ = d.Set("filename", d.Get("filename").(string))
	_ = d.Set("script_name", d.Get("script_name").(string))
	return nil
}

func deleteScript(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func createScriptFormData(fileName, scriptName string, bodyBuf *bytes.Buffer, w *multipart.Writer) error {
	scriptFile, err := os.Open(fileName)
	if err != nil {
		return err
	}

	readers := map[string]io.Reader{
		"file":       scriptFile,
		"scriptName": strings.NewReader(scriptName),
	}

	for key, r := range readers {
		var (
			fw  io.Writer
			err error
		)
		if x, ok := r.(io.Closer); ok {
			defer x.Close()
		}
		// Add an image file
		if x, ok := r.(*os.File); ok {
			fw, err = w.CreateFormFile(key, x.Name())
		} else {
			// Add other fields
			fw, err = w.CreateFormField(key)
		}
		if err != nil {
			return err
		}
		if _, err := io.Copy(fw, r); err != nil {
			return err
		}
	}

	w.Close()
	return nil
}

func scriptExistsWithName(scriptName string, meta interface{}) (bool, error) {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	scriptsApi := platformclientv2.NewScriptsApiWithConfig(sdkConfig)

	pageSize := 50
	pageNumber := 1
	data, _, err := scriptsApi.GetScripts(pageSize, pageNumber, "", scriptName, "", "", "", "", "", "")
	if err != nil {
		return false, err
	}

	if data.Entities != nil && len(*data.Entities) > 0 {
		for _, entity := range *data.Entities {
			if *entity.Name == scriptName {
				return true, nil
			}
		}
	}

	return false, nil
}

func getScriptByName(scriptName string, meta interface{}) (platformclientv2.Script, error) {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	scriptsApi := platformclientv2.NewScriptsApiWithConfig(sdkConfig)

	var script platformclientv2.Script

	pageSize := 50
	pageNumber := 1
	data, _, err := scriptsApi.GetScripts(pageSize, pageNumber, "", scriptName, "", "", "", "", "", "")
	if err != nil {
		return script, err
	}

	if data.Entities != nil && len(*data.Entities) > 0 {
		script = (*data.Entities)[0]
	} else {
		return script, fmt.Errorf("Could not find script with name %s", scriptName)
	}

	return script, nil
}
