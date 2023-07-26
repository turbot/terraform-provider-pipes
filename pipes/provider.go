package pipes

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	pipes "github.com/turbot/pipes-sdk-go"
)

// Provider
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"token": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Sets the Turbot Pipes authentication token. This is used when connecting to Turbot Pipes workspaces. You can manage your API tokens from the Settings page for your user account in Turbot Pipes.",
				DefaultFunc: schema.EnvDefaultFunc("PIPES_TOKEN", nil),
			},
			"host": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Sets the Turbot Pipes host. This is used when connecting to Turbot Pipes workspaces. The default is https://pipes.turbot.com, you only need to set this if you are connecting to a remote Turbot Pipes database that is NOT hosted in https://pipes.turbot.com, such as a dev/test instance.",
				DefaultFunc: schema.EnvDefaultFunc("PIPES_HOST", nil),
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"pipes_connection":                    resourceConnection(),
			"pipes_organization":                  resourceOrganization(),
			"pipes_organization_member":           resourceOrganizationMember(),
			"pipes_organization_workspace_member": resourceOrganizationWorkspaceMember(),
			"pipes_user_preferences":              resourceUserPreferences(),
			"pipes_workspace":                     resourceWorkspace(),
			"pipes_workspace_aggregator":          resourceWorkspaceAggregator(),
			"pipes_workspace_connection":          resourceWorkspaceConnection(),
			"pipes_workspace_mod":                 resourceWorkspaceMod(),
			"pipes_workspace_mod_variable":        resourceWorkspaceModVariable(),
			"pipes_workspace_pipeline":            resourceWorkspacePipeline(),
			"pipes_workspace_snapshot":            resourceWorkspaceSnapshot(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"pipes_organization": dataSourceOrganization(),
			"pipes_process":      dataSourceProcess(),
			"pipes_user":         dataSourceUser(),
		},

		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	config := Config{}
	if val, ok := d.GetOk("host"); ok {
		config.Host = val.(string)
	}
	if val, ok := d.GetOk("token"); ok {
		config.Token = val.(string)
	}

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	apiClient, err := CreateClient(&config, diags)
	if err != nil {
		return nil, err
	}

	log.Println("[INFO] Turbot Pipes API client initialized, now validating...", apiClient)
	return &PipesClient{
		APIClient: apiClient,
		Config:    &config,
	}, nil
}

type PipesClient struct {
	APIClient *pipes.APIClient
	Config    *Config
}

type Config struct {
	Token string
	Host  string
}

/*
precedence of credentials:
1. token set in config
2. ENV vars {PIPES_TOKEN}
*/
func CreateClient(config *Config, diags diag.Diagnostics) (*pipes.APIClient, diag.Diagnostics) {
	configuration := pipes.NewConfiguration()
	if config.Host != "" {
		parsedAPIURL, parseErr := url.Parse(config.Host)
		if parseErr != nil {
			return nil, diag.Errorf(`invalid host: %v`, parseErr)
		}
		if parsedAPIURL.Host == "" {
			return nil, diag.Errorf(`missing protocol or host : %v`, config.Host)
		}
		configuration.Servers = []pipes.ServerConfiguration{
			{
				URL: fmt.Sprintf("https://%s/api/v0", parsedAPIURL.Host),
			},
		}
	}

	var steampipeCloudToken string
	if config.Token != "" {
		steampipeCloudToken = config.Token
	} else {
		if token, ok := os.LookupEnv("STEAMPIPE_CLOUD_TOKEN"); ok {
			steampipeCloudToken = token
		} else if token, ok := os.LookupEnv("PIPES_TOKEN"); ok {
			steampipeCloudToken = token
		}
	}
	if steampipeCloudToken != "" {
		configuration.AddDefaultHeader("Authorization", fmt.Sprintf("Bearer %s", steampipeCloudToken))
		return pipes.NewAPIClient(configuration), diags
	}

	diags = append(diags, diag.Diagnostic{
		Severity: diag.Error,
		Summary:  "Unable to create Turbot Pipes client",
		Detail:   "Failed to get token to authenticate Turbot Pipes client. Please set 'token' in provider config",
	})
	return nil, diags
}
