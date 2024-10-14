package pipes

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/turbot/pipes-sdk-go"
)

// Provider
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"token": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Sets the Turbot Pipes authentication token. This is used when connecting to Turbot Pipes workspaces. You can manage your API tokens from the Settings page for your user account in Turbot Pipes.",
				DefaultFunc: schema.EnvDefaultFunc("STEAMPIPE_CLOUD_TOKEN", nil),
			},
			"host": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Sets the Turbot Pipes host. This is used when connecting to Turbot Pipes workspaces. The default is https://pipes.turbot.com, you only need to set this if you are connecting to a remote Turbot Pipes database that is NOT hosted in https://pipes.turbot.com, such as a dev/test instance.",
				DefaultFunc: schema.EnvDefaultFunc("STEAMPIPE_CLOUD_HOST", nil),
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"pipes_connection":                                resourceConnection(),
			"pipes_organization_connection":                   resourceOrganizationConnection(),
			"pipes_organization_connection_folder":            resourceOrganizationConnectionFolder(),
			"pipes_organization_connection_folder_permission": resourceOrganizationConnectionFolderPermission(),
			"pipes_organization_connection_permission":        resourceOrganizationConnectionPermission(),
			"pipes_organization_integration":                  resourceOrganizationIntegration(),
			"pipes_organization":                              resourceOrganization(),
			"pipes_organization_member":                       resourceOrganizationMember(),
			"pipes_organization_workspace_member":             resourceOrganizationWorkspaceMember(),
			"pipes_tenant_connection":                         resourceTenantConnection(),
			"pipes_tenant_connection_permission":              resourceTenantConnectionPermission(),
			"pipes_tenant_connection_folder":                  resourceTenantConnectionFolder(),
			"pipes_tenant_connection_folder_permission":       resourceTenantConnectionFolderPermission(),
			"pipes_tenant_integration":                        resourceTenantIntegration(),
			"pipes_tenant_member":                             resourceTenantMember(),
			"pipes_user_integration":                          resourceUserIntegration(),
			"pipes_user_preferences":                          resourceUserPreferences(),
			"pipes_workspace":                                 resourceWorkspace(),
			"pipes_workspace_aggregator":                      resourceWorkspaceAggregator(),
			"pipes_workspace_connection":                      resourceWorkspaceConnection(),
			"pipes_workspace_connection_folder":               resourceWorkspaceConnectionFolder(),
			"pipes_workspace_datatank":                        resourceWorkspaceDatatank(),
			"pipes_workspace_datatank_table":                  resourceWorkspaceDatatankTable(),
			"pipes_workspace_flowpipe_mod":                    resourceWorkspaceFlowpipeMod(),
			"pipes_workspace_flowpipe_mod_variable":           resourceWorkspaceFlowpipeModVariable(),
			"pipes_workspace_flowpipe_trigger":                resourceWorkspaceFlowpipeTrigger(),
			"pipes_workspace_mod":                             resourceWorkspaceMod(),
			"pipes_workspace_mod_variable":                    resourceWorkspaceModVariable(),
			"pipes_workspace_pipeline":                        resourceWorkspacePipeline(),
			"pipes_workspace_schema":                          resourceWorkspaceSchema(),
			"pipes_workspace_snapshot":                        resourceWorkspaceSnapshot(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"pipes_organization":                dataSourceOrganization(),
			"pipes_process":                     dataSourceProcess(),
			"pipes_tenant":                      dataSourceTenant(),
			"pipes_user":                        dataSourceUser(),
			"pipes_workspace_flowpipe_pipeline": dataSourceWorkspaceFlowpipePipeline(),
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
	var pipesHost string
	if config.Host != "" {
		pipesHost = config.Host
	} else {
		if host, ok := os.LookupEnv("PIPES_HOST"); ok {
			pipesHost = host
		}
	}
	if pipesHost != "" {
		parsedAPIURL, parseErr := url.Parse(pipesHost)
		if parseErr != nil {
			return nil, diag.Errorf(`invalid host: %v`, parseErr)
		}
		if parsedAPIURL.Host == "" {
			return nil, diag.Errorf(`missing protocol or host : %v`, pipesHost)
		}
		configuration.Servers = []pipes.ServerConfiguration{
			{
				URL: fmt.Sprintf("https://%s/api/v0", parsedAPIURL.Host),
			},
		}
	}

	var pipesToken string
	if config.Token != "" {
		pipesToken = config.Token
	} else {
		if token, ok := os.LookupEnv("PIPES_TOKEN"); ok {
			pipesToken = token
		}
	}
	if pipesToken != "" {
		configuration.AddDefaultHeader("Authorization", fmt.Sprintf("Bearer %s", pipesToken))
		return pipes.NewAPIClient(configuration), diags
	}

	diags = append(diags, diag.Diagnostic{
		Severity: diag.Error,
		Summary:  "Unable to create Turbot Pipes client",
		Detail:   "Failed to get token to authenticate Turbot Pipes client. Please set 'token' in provider config",
	})
	return nil, diags
}
