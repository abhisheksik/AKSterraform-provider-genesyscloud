package genesyscloud

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

func TestAccResourceWebDeploymentsConfiguration(t *testing.T) {
	t.Parallel()
	var (
		configName               = "Test Configuration " + randString(8)
		configDescription        = "Test Configuration description " + randString(32)
		updatedConfigDescription = configDescription + " Updated"
		fullResourceName         = "genesyscloud_webdeployments_configuration.basic"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: basicConfigurationResource(configName, configDescription),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fullResourceName, "name", configName),
					resource.TestCheckResourceAttr(fullResourceName, "description", configDescription),
					resource.TestMatchResourceAttr(fullResourceName, "status", regexp.MustCompile("^(Pending|Active)$")),
					resource.TestCheckResourceAttrSet(fullResourceName, "version"),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.#", "0"),
					resource.TestCheckResourceAttr(fullResourceName, "cobrowse.#", "0"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.#", "0"),
					resource.TestCheckResourceAttr(fullResourceName, "support_center.#", "0"),
				),
			},
			{
				Config: basicConfigurationResource(configName, updatedConfigDescription),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fullResourceName, "name", configName),
					resource.TestCheckResourceAttr(fullResourceName, "description", updatedConfigDescription),
					resource.TestMatchResourceAttr(fullResourceName, "status", regexp.MustCompile("^(Pending|Active)$")),
					resource.TestCheckResourceAttrSet(fullResourceName, "version"),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.#", "0"),
					resource.TestCheckResourceAttr(fullResourceName, "cobrowse.#", "0"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.#", "0"),
					resource.TestCheckResourceAttr(fullResourceName, "support_center.#", "0"),
				),
			},
			{
				ResourceName:            fullResourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"status"},
			},
		},
		CheckDestroy: verifyConfigurationDestroyed,
	})
}

func TestAccResourceWebDeploymentsConfigurationComplex(t *testing.T) {
	t.Parallel()
	var (
		configName        = "Test Configuration " + randString(8)
		configDescription = "Test Configuration description " + randString(32)
		fullResourceName  = "genesyscloud_webdeployments_configuration.complex"

		channels       = []string{strconv.Quote("Webmessaging")}
		channelsUpdate = []string{strconv.Quote("Webmessaging"), strconv.Quote("Voice")}
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: complexConfigurationResource(
					configName,
					configDescription,
					generateWebDeploymentConfigCobrowseSettings(
						TrueValue,
						TrueValue,
						channels,
						[]string{strconv.Quote("selector-one")},
						[]string{strconv.Quote("selector-one")},
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fullResourceName, "name", configName),
					resource.TestCheckResourceAttr(fullResourceName, "description", configDescription),
					resource.TestMatchResourceAttr(fullResourceName, "status", regexp.MustCompile("^(Pending|Active)$")),
					resource.TestCheckResourceAttrSet(fullResourceName, "version"),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.#", "1"),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.0.enabled", TrueValue),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.0.launcher_button.0.visibility", "OnDemand"),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.0.home_screen.0.enabled", TrueValue),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.0.home_screen.0.logo_url", ""),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.0.styles.0.primary_color", "#B0B0B0"),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.0.file_upload.0.mode.#", "2"),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.0.file_upload.0.mode.0.file_types.#", "1"),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.0.file_upload.0.mode.0.file_types.0", "image/png"),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.0.file_upload.0.mode.0.max_file_size_kb", "100"),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.0.file_upload.0.mode.1.file_types.#", "1"),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.0.file_upload.0.mode.1.file_types.0", "image/jpeg"),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.0.file_upload.0.mode.1.max_file_size_kb", "123"),
					resource.TestCheckResourceAttr(fullResourceName, "cobrowse.#", "1"),
					resource.TestCheckResourceAttr(fullResourceName, "cobrowse.0.enabled", TrueValue),
					resource.TestCheckResourceAttr(fullResourceName, "cobrowse.0.allow_agent_control", TrueValue),
					resource.TestCheckResourceAttr(fullResourceName, "cobrowse.0.channels.#", "1"),
					resource.TestCheckResourceAttr(fullResourceName, "cobrowse.0.channels.0", "Webmessaging"),
					resource.TestCheckResourceAttr(fullResourceName, "cobrowse.0.mask_selectors.#", "1"),
					resource.TestCheckResourceAttr(fullResourceName, "cobrowse.0.mask_selectors.0", "selector-one"),
					resource.TestCheckResourceAttr(fullResourceName, "cobrowse.0.readonly_selectors.#", "1"),
					resource.TestCheckResourceAttr(fullResourceName, "cobrowse.0.readonly_selectors.0", "selector-one"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.#", "1"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.enabled", TrueValue),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.excluded_query_parameters.#", "1"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.excluded_query_parameters.0", "excluded-one"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.pageview_config", "Auto"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.click_event.#", "2"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.click_event.0.selector", "first-selector"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.click_event.0.event_name", "first-click-event-name"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.click_event.1.selector", "second-selector"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.click_event.1.event_name", "second-click-event-name"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.form_track_event.#", "2"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.form_track_event.0.selector", "form-selector-1"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.form_track_event.0.form_name", "form-1"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.form_track_event.0.capture_data_on_form_abandon", TrueValue),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.form_track_event.0.capture_data_on_form_submit", FalseValue),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.form_track_event.1.selector", "form-selector-2"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.form_track_event.1.form_name", "form-3"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.form_track_event.1.capture_data_on_form_abandon", FalseValue),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.form_track_event.1.capture_data_on_form_submit", TrueValue),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.idle_event.#", "2"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.idle_event.0.event_name", "idle-event-1"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.idle_event.0.idle_after_seconds", "88"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.idle_event.1.event_name", "idle-event-2"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.idle_event.1.idle_after_seconds", "30"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.in_viewport_event.#", "2"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.in_viewport_event.0.selector", "in-viewport-selector-1"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.in_viewport_event.0.event_name", "in-viewport-event-1"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.in_viewport_event.1.selector", "in-viewport-selector-2"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.in_viewport_event.1.event_name", "in-viewport-event-2"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.scroll_depth_event.#", "2"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.scroll_depth_event.0.event_name", "scroll-depth-event-1"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.scroll_depth_event.0.percentage", "33"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.scroll_depth_event.1.event_name", "scroll-depth-event-2"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.scroll_depth_event.1.percentage", "66"),
					resource.TestCheckResourceAttr(fullResourceName, "support_center.#", "1"),
					resource.TestCheckResourceAttr(fullResourceName, "support_center.0.enabled", "true"),
					resource.TestCheckResourceAttr(fullResourceName, "support_center.0.knowledge_base_id", "dfffc742-3ba4-4363-b8e6-fbc1bea1f643"),
					resource.TestCheckResourceAttr(fullResourceName, "support_center.0.knowledge_base_uri", "/api/v2/knowledge/knowledgebases/dfffc742-3ba4-4363-b8e6-fbc1bea1f643"),
					resource.TestCheckResourceAttr(fullResourceName, "support_center.0.custom_messages.0.default_value", "Welcome to Knowledge Portal"),
					resource.TestCheckResourceAttr(fullResourceName, "support_center.0.custom_messages.1.type", "Welcome"),
					resource.TestCheckResourceAttr(fullResourceName, "support_center.0.router_type", "test"),
					resource.TestCheckResourceAttr(fullResourceName, "support_center.0.feedback_enabled", "true"),
					resource.TestCheckResourceAttr(fullResourceName, "support_center.0.enabled_categories.0.enabled_categories_id", "dfffc742-3ba4-4363-b8e6-fbc1bea1f643"),
					resource.TestCheckResourceAttr(fullResourceName, "support_center.0.enabled_categories.0.self_uri", "https://my-domain/images/my-logo.png"),
					resource.TestCheckResourceAttr(fullResourceName, "support_center.0.enabled_categories.0.image_source", ""),
					resource.TestCheckResourceAttr(fullResourceName, "support_center.0.style_setting.0.hero_style_background_color", "#5C3D5C"),
					resource.TestCheckResourceAttr(fullResourceName, "support_center.0.style_setting.0.hero_style_text_color", "#FFFFFF"),
					resource.TestCheckResourceAttr(fullResourceName, "support_center.0.style_setting.0.hero_style_image", "https://my-domain/images/my-logo.png"),
					resource.TestCheckResourceAttr(fullResourceName, "support_center.0.style_setting.0.global_style_background_color", "#FAF8F7"),
					resource.TestCheckResourceAttr(fullResourceName, "support_center.0.style_setting.0.global_style_primary_color", "#FF4F1F"),
					resource.TestCheckResourceAttr(fullResourceName, "support_center.0.style_setting.0.global_style_primary_color_dark", "#CC3F19"),
					resource.TestCheckResourceAttr(fullResourceName, "support_center.0.style_setting.0.global_style_primary_color_light", "#FFDCD2"),
					resource.TestCheckResourceAttr(fullResourceName, "support_center.0.style_setting.0.global_style_text_color", "#33302f"),
					resource.TestCheckResourceAttr(fullResourceName, "support_center.0.style_setting.0.global_style_font_family", "inherit"),
					resource.TestCheckResourceAttr(fullResourceName, "support_center.0.screens.0.type", "Home"),
					resource.TestCheckResourceAttr(fullResourceName, "support_center.0.screens.0.module_settings_type", "Search"),
					resource.TestCheckResourceAttr(fullResourceName, "support_center.0.screens.0.module_settings_enabled", "true"),
				),
			},
			{
				// Update cobrowse settings
				Config: complexConfigurationResource(
					configName,
					configDescription,
					generateWebDeploymentConfigCobrowseSettings(
						FalseValue,
						FalseValue,
						channelsUpdate,
						[]string{strconv.Quote("selector-one"), strconv.Quote("selector-two")},
						[]string{strconv.Quote("selector-one"), strconv.Quote("selector-two")},
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fullResourceName, "name", configName),
					resource.TestCheckResourceAttr(fullResourceName, "description", configDescription),
					resource.TestMatchResourceAttr(fullResourceName, "status", regexp.MustCompile("^(Pending|Active)$")),
					resource.TestCheckResourceAttrSet(fullResourceName, "version"),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.#", "1"),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.0.enabled", TrueValue),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.0.launcher_button.0.visibility", "OnDemand"),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.0.home_screen.0.enabled", TrueValue),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.0.home_screen.0.logo_url", ""),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.0.styles.0.primary_color", "#B0B0B0"),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.0.file_upload.0.mode.#", "2"),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.0.file_upload.0.mode.0.file_types.#", "1"),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.0.file_upload.0.mode.0.file_types.0", "image/png"),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.0.file_upload.0.mode.0.max_file_size_kb", "100"),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.0.file_upload.0.mode.1.file_types.#", "1"),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.0.file_upload.0.mode.1.file_types.0", "image/jpeg"),
					resource.TestCheckResourceAttr(fullResourceName, "messenger.0.file_upload.0.mode.1.max_file_size_kb", "123"),
					// resource.TestCheckResourceAttr(fullResourceName, "cobrowse.#", "1"),
					// resource.TestCheckResourceAttr(fullResourceName, "cobrowse.0.enabled", FalseValue),
					// resource.TestCheckResourceAttr(fullResourceName, "cobrowse.0.allow_agent_control", FalseValue),
					// resource.TestCheckResourceAttr(fullResourceName, "cobrowse.0.channels.#", "2"),
					// ValidateStringInArray(fullResourceName, "cobrowse.0.channels", "Webmessaging"),
					// ValidateStringInArray(fullResourceName, "cobrowse.0.channels", "Voice"),
					// resource.TestCheckResourceAttr(fullResourceName, "cobrowse.0.mask_selectors.#", "2"),
					// ValidateStringInArray(fullResourceName, "cobrowse.0.mask_selectors", "selector-one"),
					// ValidateStringInArray(fullResourceName, "cobrowse.0.mask_selectors", "selector-two"),
					// resource.TestCheckResourceAttr(fullResourceName, "cobrowse.0.readonly_selectors.#", "2"),
					// ValidateStringInArray(fullResourceName, "cobrowse.0.readonly_selectors", "selector-one"),
					// ValidateStringInArray(fullResourceName, "cobrowse.0.readonly_selectors", "selector-two"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.#", "1"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.enabled", TrueValue),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.excluded_query_parameters.#", "1"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.excluded_query_parameters.0", "excluded-one"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.pageview_config", "Auto"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.click_event.#", "2"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.click_event.0.selector", "first-selector"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.click_event.0.event_name", "first-click-event-name"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.click_event.1.selector", "second-selector"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.click_event.1.event_name", "second-click-event-name"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.form_track_event.#", "2"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.form_track_event.0.selector", "form-selector-1"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.form_track_event.0.form_name", "form-1"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.form_track_event.0.capture_data_on_form_abandon", TrueValue),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.form_track_event.0.capture_data_on_form_submit", FalseValue),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.form_track_event.1.selector", "form-selector-2"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.form_track_event.1.form_name", "form-3"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.form_track_event.1.capture_data_on_form_abandon", FalseValue),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.form_track_event.1.capture_data_on_form_submit", TrueValue),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.idle_event.#", "2"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.idle_event.0.event_name", "idle-event-1"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.idle_event.0.idle_after_seconds", "88"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.idle_event.1.event_name", "idle-event-2"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.idle_event.1.idle_after_seconds", "30"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.in_viewport_event.#", "2"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.in_viewport_event.0.selector", "in-viewport-selector-1"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.in_viewport_event.0.event_name", "in-viewport-event-1"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.in_viewport_event.1.selector", "in-viewport-selector-2"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.in_viewport_event.1.event_name", "in-viewport-event-2"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.scroll_depth_event.#", "2"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.scroll_depth_event.0.event_name", "scroll-depth-event-1"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.scroll_depth_event.0.percentage", "33"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.scroll_depth_event.1.event_name", "scroll-depth-event-2"),
					resource.TestCheckResourceAttr(fullResourceName, "journey_events.0.scroll_depth_event.1.percentage", "66"),
					resource.TestCheckResourceAttr(fullResourceName, "support_center.#", "1"),
					resource.TestCheckResourceAttr(fullResourceName, "support_center.0.enabled", "true"),
					resource.TestCheckResourceAttr(fullResourceName, "support_center.0.knowledge_base_id", "dfffc742-3ba4-4363-b8e6-fbc1bea1f643"),
					resource.TestCheckResourceAttr(fullResourceName, "support_center.0.knowledge_base_uri", "/api/v2/knowledge/knowledgebases/dfffc742-3ba4-4363-b8e6-fbc1bea1f643"),
					resource.TestCheckResourceAttr(fullResourceName, "support_center.0.custom_messages.0.default_value", "Welcome to Knowledge Portal"),
					resource.TestCheckResourceAttr(fullResourceName, "support_center.0.custom_messages.1.type", "Welcome"),
					resource.TestCheckResourceAttr(fullResourceName, "support_center.0.router_type", "test"),
					resource.TestCheckResourceAttr(fullResourceName, "support_center.0.feedback_enabled", "true"),
					resource.TestCheckResourceAttr(fullResourceName, "support_center.0.enabled_categories.0.enabled_categories_id", "dfffc742-3ba4-4363-b8e6-fbc1bea1f643"),
					resource.TestCheckResourceAttr(fullResourceName, "support_center.0.enabled_categories.0.self_uri", "https://my-domain/images/my-logo.png"),
					resource.TestCheckResourceAttr(fullResourceName, "support_center.0.enabled_categories.0.image_source", ""),
					resource.TestCheckResourceAttr(fullResourceName, "support_center.0.style_setting.0.hero_style_background_color", "#5C3D5C"),
					resource.TestCheckResourceAttr(fullResourceName, "support_center.0.style_setting.0.hero_style_text_color", "#FFFFFF"),
					resource.TestCheckResourceAttr(fullResourceName, "support_center.0.style_setting.0.hero_style_image", "https://my-domain/images/my-logo.png"),
					resource.TestCheckResourceAttr(fullResourceName, "support_center.0.style_setting.0.global_style_background_color", "#FAF8F7"),
					resource.TestCheckResourceAttr(fullResourceName, "support_center.0.style_setting.0.global_style_primary_color", "#FF4F1F"),
					resource.TestCheckResourceAttr(fullResourceName, "support_center.0.style_setting.0.global_style_primary_color_dark", "#CC3F19"),
					resource.TestCheckResourceAttr(fullResourceName, "support_center.0.style_setting.0.global_style_primary_color_light", "#FFDCD2"),
					resource.TestCheckResourceAttr(fullResourceName, "support_center.0.style_setting.0.global_style_text_color", "#33302f"),
					resource.TestCheckResourceAttr(fullResourceName, "support_center.0.style_setting.0.global_style_font_family", "inherit"),
					resource.TestCheckResourceAttr(fullResourceName, "support_center.0.screens.0.type", "Home"),
					resource.TestCheckResourceAttr(fullResourceName, "support_center.0.screens.0.module_settings_type", "Search"),
					resource.TestCheckResourceAttr(fullResourceName, "support_center.0.screens.0.module_settings_enabled", "true"),
				),
			},
			{
				ResourceName:            fullResourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"status"},
			},
		},
		CheckDestroy: verifyConfigurationDestroyed,
	})
}

func basicConfigurationResource(name, description string) string {
	return fmt.Sprintf(`
	resource "genesyscloud_webdeployments_configuration" "basic" {
		name             = "%s"
		description      = "%s"
		languages        = [ "en-us", "ja" ]
		default_language = "en-us"
	}
	`, name, description)
}

func complexConfigurationResource(name, description string, nestedBlocks ...string) string {
	return fmt.Sprintf(`
	resource "genesyscloud_webdeployments_configuration" "complex" {
		name = "%s"
		description = "%s"
		languages = [ "en-us", "ja" ]
		default_language = "en-us"
		messenger {
			enabled = true
			launcher_button {
				visibility = "OnDemand"
			}
			home_screen {
				enabled = true
				logo_url = ""
			}
			styles {
				primary_color = "#B0B0B0"
			}
			file_upload {
				mode {
					file_types = [ "image/png" ]
					max_file_size_kb = 100
				}
				mode {
					file_types = [ "image/jpeg" ]
					max_file_size_kb = 123
				}
			}
		}
		journey_events {
			enabled = true
			excluded_query_parameters = [ "excluded-one" ]

			pageview_config = "Auto"

			click_event {
				selector = "first-selector"
				event_name = "first-click-event-name"
			}
			click_event {
				selector = "second-selector"
				event_name = "second-click-event-name"
			}

			form_track_event {
				selector = "form-selector-1"
				form_name = "form-1"
				capture_data_on_form_abandon = true
				capture_data_on_form_submit = false
			}

			form_track_event {
				selector = "form-selector-2"
				form_name = "form-3"
				capture_data_on_form_abandon = false
				capture_data_on_form_submit = true
			}

			idle_event {
				event_name = "idle-event-1"
				idle_after_seconds = 88
			}

			idle_event {
				event_name = "idle-event-2"
				idle_after_seconds = 30
			}

			in_viewport_event {
				selector = "in-viewport-selector-1"
				event_name = "in-viewport-event-1"
			}

			in_viewport_event {
				selector = "in-viewport-selector-2"
				event_name = "in-viewport-event-2"
			}

			scroll_depth_event {
				event_name = "scroll-depth-event-1"
				percentage = 33
			}

			scroll_depth_event {
				event_name = "scroll-depth-event-2"
				percentage = 66
			}
		}

		support_center {
			enabled = true
			knowledge_base_id = "dfffc742-3ba4-4363-b8e6-fbc1bea1f643"
			knowledge_base_uri = "/api/v2/knowledge/knowledgebases/dfffc742-3ba4-4363-b8e6-fbc1bea1f643"
			custom_messages {
				default_value = "Welcome to Knowledge Portal"
				type ="Welcome"
			}
			router_type = ""
			feedback_enabled = ""
			enabled_categories {
				enabled_categories_id = "dfffc742-3ba4-4363-b8e6-fbc1bea1f643"
				self_uri = "https://example.com/image"
				image_source = "https://example.com/image"
			}
			style_setting {
				hero_style_background_color = "#5C3D5C"
				hero_style_text_color = "#FFFFFF"
				hero_style_image = "https://example.com/image"
				global_style_background_color = "#FAF8F7"
				global_style_primary_color = "#FF4F1F"
				global_style_primary_color_dark = "#CC3F19"
				global_style_primary_color_light = "#FFDCD2"
				global_style_text_color = "#33302f"
				global_style_font_family = "inherit"
			}
			screens {
				type = "Home"
				module_settings_type = "Search"
				module_settings_enabled = true
				module_settings_compact_category_module_template = false
				module_settings_detailed_category_module_template = false

				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			}
		}
		%s
	}
	`, name, description, strings.Join(nestedBlocks, "\n"))
}

func generateWebDeploymentConfigCobrowseSettings(cbEnabled, cbAllowAgentControl string, cbChannels []string, cbMaskSelectors []string, cbReadonlySelectors []string) string {
	return fmt.Sprintf(`
	cobrowse {
		enabled = %s
		allow_agent_control = %s
		channels = [ %s ]
		mask_selectors = [ %s ]
		readonly_selectors = [ %s ]
	}
`, cbEnabled, cbAllowAgentControl, strings.Join(cbChannels, ", "), strings.Join(cbMaskSelectors, ", "), strings.Join(cbReadonlySelectors, ", "))
}

func generateSupportCenterSettings(enabled, knowledgeBaseID, knowledgeBaseURI string) string {
	return fmt.Sprintf(`
		support_center {
			enabled = %s
			knowledge_base_id = "%s"
			knowledge_base_uri = "%s"
			custom_messages {
				default_value = "Welcome to Knowledge Portal"
				type = "Welcome"
			}
			router_type = ""
			feedback_enabled = ""
			enabled_categories {
				enabled_categories_id = "dfffc742-3ba4-4363-b8e6-fbc1bea1f643"
				self_uri = "https://example.com/image"
				image_source = "https://example.com/image"
			}
			style_setting {
				hero_style_background_color = "#5C3D5C"
				hero_style_text_color = "#FFFFFF"
				hero_style_image = "https://example.com/image"
				global_style_background_color = "#FAF8F7"
				global_style_primary_color = "#FF4F1F"
				global_style_primary_color_dark = "#CC3F19"
				global_style_primary_color_light = "#FFDCD2"
				global_style_text_color = "#33302f"
				global_style_font_family = "inherit"
			}
			screens {
				type = "Home"
				module_settings_type = "Search"
				module_settings_enabled = true
				module_settings_compact_category_module_template = false
				module_settings_detailed_category_module_template = false
			}
		}
	`, enabled, knowledgeBaseID, knowledgeBaseURI)
}

func verifyConfigurationDestroyed(state *terraform.State) error {
	api := platformclientv2.NewWebDeploymentsApi()

	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_webdeployments_configuration" {
			continue
		}

		_, response, err := api.GetWebdeploymentsConfigurationVersionsDraft(rs.Primary.ID)

		if IsStatus404(response) {
			continue
		}

		if err != nil {
			return fmt.Errorf("Unexpected error while checking that configuration has been destroyed: %s", err)
		}

		return fmt.Errorf("Configuration %s still exists when it was expected to have been destroyed", rs.Primary.ID)
	}

	return nil
}
