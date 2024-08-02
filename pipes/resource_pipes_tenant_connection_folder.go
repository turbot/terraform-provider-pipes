package pipes

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	pipes "github.com/turbot/pipes-sdk-go"
)

func resourceTenantConnectionFolder() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTenantConnectionFolderCreate,
		ReadContext:   resourceTenantConnectionFolderRead,
		UpdateContext: resourceTenantConnectionFolderUpdate,
		DeleteContext: resourceTenantConnectionFolderDelete,
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
		},
	}
}

func resourceTenantConnectionFolderCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var title, parentId string
	var err error

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

	resp, r, err = client.APIClient.TenantConnectionFolders.Create(ctx).Request(req).Execute()
	if err != nil {
		return diag.Errorf("resourceTenantConnectionFolderCreate. Create connection api error  %v", decodeResponse(r))
	}

	d.Set("connection_folder_id", resp.Id)
	d.Set("tenant_id", resp.TenantId)
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
	// The connection folder is being created at a custom tenant level
	// The id would be of format "TenantId/ConnectionFolderId"
	d.SetId(fmt.Sprintf("%s/%s", resp.TenantId, resp.Id))

	return diags
}

func resourceTenantConnectionFolderRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*PipesClient)

	// Warning or errors can be collected in a slice type
	var connectionFolderId, tenantId string
	var diags diag.Diagnostics

	// Its a tenant level connection so the id would be of format "TenantId/ConnectionFolderId"
	ids := strings.Split(d.Id(), "/")
	if len(ids) == 2 {
		tenantId = ids[0]
		connectionFolderId = ids[1]
	}

	if tenantId == "" {
		return diag.Errorf("resourceTenantConnectionFolderRead. Tenant information not present.")
	}
	if connectionFolderId == "" {
		return diag.Errorf("resourceTenantConnectionFolderRead. Connection folder id not present.")
	}

	resp, r, err := client.APIClient.TenantConnectionFolders.Get(context.Background(), connectionFolderId).Execute()
	if err != nil {
		if r.StatusCode == 404 {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  fmt.Sprintf("Connection Folder (%s) not found", connectionFolderId),
			})
			d.SetId("")
			return diags
		}
		return diag.Errorf("resourceTenantConnectionFolderRead. Get connection error: %v", decodeResponse(r))
	}

	d.Set("connection_folder_id", resp.Id)
	d.Set("tenant_id", resp.TenantId)
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
	// The connection folder is being created at a custom tenant level
	// The id would be of format "TenantId/ConnectionFolderId"
	d.SetId(fmt.Sprintf("%s/%s", resp.TenantId, resp.Id))

	return diags
}

func resourceTenantConnectionFolderUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*PipesClient)

	connectionFolderId := strings.Split(d.Id(), "/")[1]

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	_, new := d.GetChange("title")
	if new.(string) == "" {
		return diag.Errorf("title must be configured for a connection folder")
	}

	newTitle := new.(string)

	req := pipes.UpdateConnectionFolderRequest{
		Title: &newTitle,
	}
	if ok := d.HasChange("parent_id"); ok {
		if value, ok := d.GetOk("parent_id"); ok {
			req.SetParentId(value.(string))
		}
	}

	resp, r, err := client.APIClient.TenantConnectionFolders.Update(context.Background(), connectionFolderId).Request(req).Execute()
	if err != nil {
		return diag.Errorf("resourceTenantConnectionFolderUpdate. Update connection error: %v", decodeResponse(r))
	}

	d.Set("connection_folder_id", resp.Id)
	d.Set("tenant_id", resp.TenantId)
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
	// The connection folder is being created at a custom tenant level
	// The id would be of format "TenantId/ConnectionFolderId"
	d.SetId(fmt.Sprintf("%s/%s", resp.TenantId, resp.Id))

	return diags
}

func resourceTenantConnectionFolderDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*PipesClient)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	connectionFolderId := strings.Split(d.Id(), "/")[1]

	_, r, err := client.APIClient.TenantConnectionFolders.Delete(ctx, connectionFolderId).Execute()
	if err != nil {
		return diag.Errorf("resourceTenantConnectionFolderDelete. Delete connection error:	%v", decodeResponse(r))
	}

	// clear the id to show we have deleted
	d.SetId("")

	return diags
}
