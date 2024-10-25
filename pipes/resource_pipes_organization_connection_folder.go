package pipes

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/turbot/pipes-sdk-go"
)

func resourceOrganizationConnectionFolder() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOrganizationConnectionFolderCreate,
		ReadContext:   resourceOrganizationConnectionFolderRead,
		UpdateContext: resourceOrganizationConnectionFolderUpdate,
		DeleteContext: resourceOrganizationConnectionFolderDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"connection_folder_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"tenant_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"organization_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"title": {
				Type:     schema.TypeString,
				Required: true,
			},
			"type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"parent_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"integration_resource_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"integration_resource_identifier": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"integration_resource_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"integration_resource_path": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"managed_by_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"trunk": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeMap},
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_by": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"updated_by": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"version_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"organization": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceOrganizationConnectionFolderCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var orgHandle, title, parentId string
	var err error

	// Get details about the organization where the connection folder would be created
	if val, ok := d.GetOk("organization"); ok {
		orgHandle = val.(string)
	}

	// Get title and parent_id for the connection folder to be created
	if value, ok := d.GetOk("title"); ok {
		title = value.(string)
	}
	if value, ok := d.GetOk("parent_id"); ok {
		parentId = value.(string)
	}

	req := pipes.CreateConnectionFolderRequest{
		Title: title,
	}

	// Pass the parent_id if its set
	if parentId != "" {
		req.SetParentId(parentId)
	}

	client := meta.(*PipesClient)
	var resp pipes.Connection
	var r *http.Response

	resp, r, err = client.APIClient.OrgConnectionFolders.Create(ctx, orgHandle).Request(req).Execute()
	if err != nil {
		return diag.Errorf("resourceOrganizationConnectionFolderCreate. Create connection api error  %v", decodeResponse(r))
	}

	d.Set("connection_folder_id", resp.Id)
	d.Set("tenant_id", resp.TenantId)
	d.Set("organization_id", resp.IdentityId)
	d.Set("title", resp.Title)
	d.Set("type", resp.Type)
	d.Set("parent_id", resp.ParentId)
	d.Set("integration_resource_name", resp.IntegrationResourceName)
	d.Set("integration_resource_identifier", resp.IntegrationResourceIdentifier)
	d.Set("integration_resource_type", resp.IntegrationResourceType)
	d.Set("integration_resource_path", resp.IntegrationResourcePath)
	d.Set("managed_by_id", resp.ManagedById)
	d.Set("trunk", resp.Trunk)
	d.Set("created_at", resp.CreatedAt)
	d.Set("updated_at", resp.UpdatedAt)
	if resp.CreatedBy != nil {
		d.Set("created_by", resp.CreatedBy.Handle)
	}
	if resp.UpdatedBy != nil {
		d.Set("updated_by", resp.UpdatedBy.Handle)
	}
	d.Set("version_id", resp.VersionId)
	d.Set("organization", orgHandle)
	// The id would be of format "OrganizationHandle/ConnectionFolderId"
	d.SetId(fmt.Sprintf("%s/%s", orgHandle, resp.Id))

	return diags
}

func resourceOrganizationConnectionFolderRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*PipesClient)

	// Warning or errors can be collected in a slice type
	var connectionFolderId, orgHandle string
	var diags diag.Diagnostics

	// Id would be of format "OrganizationHandle/ConnectionFolderId"
	ids := strings.Split(d.Id(), "/")
	if len(ids) == 2 {
		orgHandle = ids[0]
		connectionFolderId = ids[1]
	}

	if orgHandle == "" {
		return diag.Errorf("resourceOrganizationConnectionFolderRead. Organization information not present.")
	}
	if connectionFolderId == "" {
		return diag.Errorf("resourceOrganizationConnectionFolderRead. Connection folder id not present.")
	}

	resp, r, err := client.APIClient.OrgConnectionFolders.Get(context.Background(), orgHandle, connectionFolderId).Execute()
	if err != nil {
		if r.StatusCode == 404 {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  fmt.Sprintf("Connection Folder (%s) not found", connectionFolderId),
			})
			d.SetId("")
			return diags
		}
		return diag.Errorf("resourceOrganizationConnectionFolderRead. Get connection error: %v", decodeResponse(r))
	}

	d.Set("connection_folder_id", resp.Id)
	d.Set("tenant_id", resp.TenantId)
	d.Set("organization_id", resp.IdentityId)
	d.Set("title", resp.Title)
	d.Set("type", resp.Type)
	d.Set("parent_id", resp.ParentId)
	d.Set("integration_resource_name", resp.IntegrationResourceName)
	d.Set("integration_resource_identifier", resp.IntegrationResourceIdentifier)
	d.Set("integration_resource_type", resp.IntegrationResourceType)
	d.Set("integration_resource_path", resp.IntegrationResourcePath)
	d.Set("managed_by_id", resp.ManagedById)
	d.Set("trunk", resp.Trunk)
	d.Set("created_at", resp.CreatedAt)
	d.Set("updated_at", resp.UpdatedAt)
	if resp.CreatedBy != nil {
		d.Set("created_by", resp.CreatedBy.Handle)
	}
	if resp.UpdatedBy != nil {
		d.Set("updated_by", resp.UpdatedBy.Handle)
	}
	d.Set("version_id", resp.VersionId)
	d.Set("organization", orgHandle)
	// The id would be of format "OrganizationHandle/ConnectionFolderId"
	d.SetId(fmt.Sprintf("%s/%s", orgHandle, resp.Id))

	return diags
}

func resourceOrganizationConnectionFolderUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*PipesClient)

	var connectionFolderId, orgHandle string
	ids := strings.Split(d.Id(), "/")
	if len(ids) == 2 {
		orgHandle = ids[0]
		connectionFolderId = ids[1]
	}

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	_, n := d.GetChange("title")
	newTitle := n.(string)
	if newTitle == "" {
		return diag.Errorf("title must be configured for a connection folder")
	}

	req := pipes.UpdateConnectionFolderRequest{
		Title: &newTitle,
	}
	if ok := d.HasChange("parent_id"); ok {
		if value, ok := d.GetOk("parent_id"); ok {
			req.SetParentId(value.(string))
		}
	}

	resp, r, err := client.APIClient.OrgConnectionFolders.Update(context.Background(), orgHandle, connectionFolderId).Request(req).Execute()
	if err != nil {
		return diag.Errorf("resourceOrganizationConnectionFolderUpdate. Update connection error: %v", decodeResponse(r))
	}

	d.Set("connection_folder_id", resp.Id)
	d.Set("tenant_id", resp.TenantId)
	d.Set("organization_id", resp.IdentityId)
	d.Set("title", resp.Title)
	d.Set("type", resp.Type)
	d.Set("parent_id", resp.ParentId)
	d.Set("integration_resource_name", resp.IntegrationResourceName)
	d.Set("integration_resource_identifier", resp.IntegrationResourceIdentifier)
	d.Set("integration_resource_type", resp.IntegrationResourceType)
	d.Set("integration_resource_path", resp.IntegrationResourcePath)
	d.Set("managed_by_id", resp.ManagedById)
	d.Set("trunk", resp.Trunk)
	d.Set("created_at", resp.CreatedAt)
	d.Set("updated_at", resp.UpdatedAt)
	if resp.CreatedBy != nil {
		d.Set("created_by", resp.CreatedBy.Handle)
	}
	if resp.UpdatedBy != nil {
		d.Set("updated_by", resp.UpdatedBy.Handle)
	}
	d.Set("version_id", resp.VersionId)
	d.Set("organization", orgHandle)
	// The id would be of format "OrganizationHandle/ConnectionFolderId"
	d.SetId(fmt.Sprintf("%s/%s", orgHandle, resp.Id))

	return diags
}

func resourceOrganizationConnectionFolderDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*PipesClient)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var connectionFolderId, orgHandle string
	ids := strings.Split(d.Id(), "/")
	if len(ids) == 2 {
		orgHandle = ids[0]
		connectionFolderId = ids[1]
	}

	_, r, err := client.APIClient.OrgConnectionFolders.Delete(ctx, orgHandle, connectionFolderId).Execute()
	if err != nil {
		return diag.Errorf("resourceOrganizationConnectionFolderDelete. Delete connection error:	%v", decodeResponse(r))
	}

	// clear the id to show we have deleted
	d.SetId("")

	return diags
}
