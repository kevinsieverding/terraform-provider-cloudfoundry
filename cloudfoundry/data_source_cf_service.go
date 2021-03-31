package cloudfoundry

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceService() *schema.Resource {

	return &schema.Resource{

		ReadContext: dataSourceServiceRead,

		Schema: map[string]*schema.Schema{

			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"space": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"service_broker_guid": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"service_broker_name": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"service_plans": &schema.Schema{
				Type:     schema.TypeMap,
				Computed: true,
			},
		},
	}
}

func dataSourceServiceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	session := meta.(*managers.Session)

	name := d.Get("name").(string)
	space := d.Get("space").(string)
	serviceBrokerGUID := d.Get("service_broker_guid").(string)

	filters := []ccv2.Filter{
		ccv2.FilterEqual(constant.LabelFilter, name),
	}

	if serviceBrokerGUID != "" {
		filters = append(filters, ccv2.FilterEqual(constant.ServiceBrokerGUIDFilter, serviceBrokerGUID))
	}
	services, _, err := session.ClientV2.GetServices(filters...)
	if err != nil {
		return diag.FromErr(err)
	}
	if len(services) == 0 {
		return diag.FromErr(NotFound)
	}

	service := services[0]
	if space != "" {
		for _, svc := range services {
			brk, _, err := session.ClientV2.GetServiceBroker(svc.ServiceBrokerGUID)
			if err != nil {
				return err
			}
			if brk.SpaceGUID == space {
				service = svc
				break
			}
		}

	}
	d.SetId(service.GUID)
	if serviceBrokerGUID == "" {
		d.Set("service_broker_name", service.ServiceBrokerName)
	}

	servicePlans, _, err := session.ClientV2.GetServicePlans(ccv2.FilterEqual(constant.ServiceGUIDFilter, service.GUID))
	if err != nil {
		return diag.FromErr(err)
	}

	servicePlansTf := make(map[string]interface{})
	for _, sp := range servicePlans {
		servicePlansTf[strings.Replace(sp.Name, ".", "_", -1)] = sp.GUID
	}
	d.Set("service_plans", servicePlansTf)

	return nil
}
