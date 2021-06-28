package aws

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/iam/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/iam/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func dataSourceAwsIAMAssumedRoleSource() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsIAMAssumedRoleSourceRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Required: true,
			},
			"role_path": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"role_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"source_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"session_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsIAMAssumedRoleSourceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iamconn

	arn := d.Get("arn").(string)

	d.SetId(arn)

	roleName := ""
	sessionName := ""
	var err error

	if roleName, sessionName = roleNameSessionFromARN(arn); roleName == "" {
		d.Set("source_arn", arn)
		d.Set("session_name", "")
		d.Set("role_name", "")
		d.Set("role_path", "")

		return nil
	}

	var role *iam.Role

	err = resource.Retry(waiter.PropagationTimeout, func() *resource.RetryError {
		var err error

		role, err = finder.RoleByName(conn, roleName)

		if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		role, err = finder.RoleByName(conn, roleName)
	}

	if err != nil {
		return fmt.Errorf("unable to get role (%s): %w", roleName, err)
	}

	if role == nil || role.Arn == nil {
		return fmt.Errorf("empty role returned (%s)", roleName)
	}

	d.Set("session_name", sessionName)
	d.Set("role_name", roleName)
	d.Set("source_arn", role.Arn)
	d.Set("role_path", role.Path)

	return nil
}

// roleNameSessionFromARN returns the role and session names in an ARN if any.
// Otherwise, it returns empty strings.
func roleNameSessionFromARN(rawARN string) (string, string) {
	parsedARN, err := arn.Parse(rawARN)

	if err != nil {
		return "", ""
	}

	parts := strings.Split(parsedARN.Resource, "/")

	reAssume := regexp.MustCompile(`^assumed-role/.{1,}/.{2,}`)
	reRole := regexp.MustCompile(`^role/.{1,}`)

	if reAssume.MatchString(parsedARN.Resource) && parsedARN.Service != "sts" {
		return "", ""
	}

	if reRole.MatchString(parsedARN.Resource) && parsedARN.Service != "iam" {
		return "", ""
	}

	if !reAssume.MatchString(parsedARN.Resource) && !reRole.MatchString(parsedARN.Resource) {
		return "", ""
	}

	if reRole.MatchString(parsedARN.Resource) && len(parts) > 1 {
		return parts[len(parts)-1], ""
	}

	if len(parts) < 3 {
		return "", ""
	}

	return parts[len(parts)-2], parts[len(parts)-1]
}