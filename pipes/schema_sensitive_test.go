package pipes

import (
	"regexp"
	"slices"
	"sort"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// credentialAttribute matches attribute names that can hold a credential. Terraform prints an
// attribute's value in plan and apply output unless its schema sets Sensitive, so every match
// must set it. A connection or integration config can hold the access keys, client secrets or
// tokens of the system it connects to. See GHSA-f5px-m78q-8xxj.
var credentialAttribute = regexp.MustCompile(`^config$|secret|password|private_key|token`)

// notCredentials match credentialAttribute but hold no credential.
var notCredentials = map[string]bool{
	"pipes_tenant.token_min_issued_at":           true,
	"data.pipes_tenant.token_min_issued_at":      true,
	"pipes_tenant_settings.max_token_expiration": true,
}

type schemaAttribute struct {
	path   string
	name   string
	schema *schema.Schema
}

func providerAttributes() (resources, dataSources, provider []schemaAttribute) {
	p := Provider()
	provider = walkAttributes("provider", p.Schema)
	for name, r := range p.ResourcesMap {
		resources = append(resources, walkAttributes(name, r.Schema)...)
	}
	for name, r := range p.DataSourcesMap {
		dataSources = append(dataSources, walkAttributes("data."+name, r.Schema)...)
	}
	return resources, dataSources, provider
}

func walkAttributes(prefix string, attrs map[string]*schema.Schema) []schemaAttribute {
	var out []schemaAttribute
	for name, s := range attrs {
		path := prefix + "." + name
		out = append(out, schemaAttribute{path: path, name: name, schema: s})
		if nested, ok := s.Elem.(*schema.Resource); ok {
			out = append(out, walkAttributes(path, nested.Schema)...)
		}
	}
	return out
}

func TestCredentialAttributesAreSensitive(t *testing.T) {
	resources, dataSources, provider := providerAttributes()
	all := slices.Concat(resources, dataSources, provider)

	var found, missing []string
	for _, a := range all {
		if !credentialAttribute.MatchString(a.name) || notCredentials[a.path] {
			continue
		}
		found = append(found, a.path)
		if !a.schema.Sensitive {
			missing = append(missing, a.path)
		}
	}

	sort.Strings(missing)
	for _, path := range missing {
		t.Errorf("%s can hold a credential but is not Sensitive", path)
	}

	// Guard the scan itself: if these stop matching, the walk is broken and the test proves nothing.
	for _, want := range []string{
		"provider.token",
		"pipes_connection.config",
		"pipes_organization_connection.config",
		"pipes_tenant_connection.config",
		"pipes_workspace_connection.config",
		"pipes_organization_integration.config",
		"pipes_tenant_integration.config",
		"pipes_user_integration.config",
		"data.pipes_organization_integration.config",
		"data.pipes_tenant_integration.config",
		"data.pipes_user_integration.config",
	} {
		if !slices.Contains(found, want) {
			t.Errorf("scan did not reach %s", want)
		}
	}
}

// A data source attribute shows its value in plan output just as a resource attribute does, so
// an attribute name that a resource marks Sensitive must be Sensitive on every data source too.
func TestDataSourceAttributesAreSensitiveWhereResourcesAre(t *testing.T) {
	resources, dataSources, _ := providerAttributes()

	sensitiveOnResource := map[string]string{}
	for _, a := range resources {
		if a.schema.Sensitive {
			sensitiveOnResource[a.name] = a.path
		}
	}

	var missing []string
	for _, a := range dataSources {
		if resourcePath, ok := sensitiveOnResource[a.name]; ok && !a.schema.Sensitive {
			missing = append(missing, a.path+" (Sensitive on "+resourcePath+")")
		}
	}

	sort.Strings(missing)
	for _, m := range missing {
		t.Errorf("%s is not Sensitive", m)
	}
}
