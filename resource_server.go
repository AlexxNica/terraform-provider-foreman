package main

import (
	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/schema"
  "strings"
	//"reflect"
	"fmt"
	"net/http"
	"encoding/json"
	"io/ioutil"
	//"os"
	"bytes"
	//"strconv"
	"errors"
)

//reqHost is required to wrap the host for the foreman-api
type reqHost struct {
	Lhost host `json:"host,omitempty"`
}

//setup param archetype for host params attributes
type params_archetype struct {
	Name	string	`json:"name,omitempty"`
	Value	string	`json:"value,omitempty"`
}
//host_parameters_attributes parameters
type host_parameters_attributes	struct {
  Parameters		[]params_archetype	`json:",omitempty"`
}
//interfaces_attributes parameters
type interfaces_attributes	struct	{
	Mac 								string	`json:"mac,omitempty"`
	Ip 									string	`json:"ip,omitempty"`
	Type 								string	`json:"type,omitempty"`
	Name 								string	`json:"name,omitempty"`
	Subnet_id 					int			`json:"subnet_id,omitempty"`
	Domain_id 					int			`json:"domain_id,omitempty"`
	Identifier 					string	`json:"identifier,omitempty"`
	Managed 						bool		`json:"managed,omitempty"`
	Primary 						bool		`json:"primary,omitempty"`
	Provision 					bool		`json:"provision,omitempty"`
	Username 						string	`json:"username,omitempty"`//only for bmc
	Password 						string	`json:"password,omitempty"` //only for bmc
	Provider 						string	`json:"provider,omitempty"` //only accepted IPMI
	Virtual 						bool		`json:"virtual,omitempty"`
	Tag 								string	`json:"tag,omitempty"`
	Attached_to 				string	`json:"attached_to,omitempty"`
	Mode 								string	`json:"mode,omitempty"` // with validations
	Attached_devices 		[]string `json:"attached_devices,omitempty"`
	Bond_options 				string	`json:"bond_options,omitempty"`
	//These are special attributes for hypervisor like vlan and whatnot
	Lcompute_attributes ifcompute_attributes `json:"compute_attributes,omitempty"`
}
//These are actual compute instance attributes
type compute_attributes	struct {
	Cpus 			string	`json:"cpus,omitempty"`
	Start 		string	`json:"start,omitempty"`
	Cluster 	string	`json:"cluster,omitempty"`
	Memory_mb string	`json:"memory_mb,omitempty"`
	Guest_id 	string	`json:"guest_id,omitempty"`
	//This needs to be a struct map or the foreman API will kick back the JSON
	Lvolumes_attributes	map[string]volumes_attributes	`json:"volumes_attributes,omitempty"`
}

//struct for nested interface compute attributes
type ifcompute_attributes struct {
	Network	string	`json:"network,omitempty"`
	Type		string	`json:"type,omitempty"`
}
//These are things like which datastore or cluster the virtual disks need to live on
type volumes_attributes struct {
	Name		  string	`json:"name,omitempty"`
	Size_gb	  int			`json:"size_gb,omitempty"`
	_delete	  string	`json:",omitempty"`
	Datastore	string	`json:"datastore,omitempty"`
}
//This is the main host struct instance that later gets wrapped in reqHost for JSON/foreman API reasons
type host struct {
  Name									string	`json:"name,omitempty"`
  Environment_id				string	`json:"environment_id,omitempty"`
  Ip										string	`json:"ip,omitempty"`
  Mac										string	`json:"mac,omitempty"`
  Architecture_id				int			`json:"architecture_id,omitempty"`
  Domain_id							int			`json:"domain_id,omitempty"`
  Realm_id							int			`json:"realm_id,omitempty"`
  Puppet_proxy_id				int			`json:"puppet_proxy_id,omitempty"`
  Puppetclass_ids				[]int		`json:"puppetclass_ids,omitempty"`
  Operatingsystem_id		string	`json:"operatingsystem_id,omitempty"`
  Medium_id							string	`json:"medium_id,omitempty"`
  Ptable_id							int			`json:"ptable_id,omitempty"`
  Subnet_id							int			`json:"subnet_id,omitempty"`
  Compute_resource_id		int			`json:"compute_resource_id,omitempty"`
  Root_pass							string	`json:"root_pass,omitempty"`
  Model_id							int			`json:"model_id,omitempty"`
  Hostgroup_id					int			`json:"hostgroup_id,omitempty"`
  Puppet_ca_proxy_id		int			`json:"puppet_ca_proxy_id,omitempty"`
  Image_id							int			`json:"image_id,omitempty"`
  Build									bool		`json:"build,omitempty"`
  Enabled								bool		`json:"enabled,omitempty"`
  Provision_method			string	`json:"provision_method,omitempty"`
  Managed								bool		`json:"managed,omitempty"`
	Lcompute_attributes		compute_attributes	`json:"compute_attributes,omitempty"`
	Owner_id							int			`json:"owner_id,omitempty"`
	Owner_type						string	`json:"owner_type,omitempty"` // must be either User or Usergroup
  Progress_report_id		string	`json:"progress_report_id,omitempty"`
  Comment								string	`json:"comment,omitempty"`
  Capabilities					string	`json:"capabilities,omitempty"`
  Compute_profile_id		int			`json:"compute_profile_id,omitempty"`
	Lhost_parameters_attributes []host_parameters_attributes	`json:"host_parameters_attributes,omitempty"`
  Linterfaces_attributes	[]interfaces_attributes	`json:"interfaces_attributes,omitempty"`
}
//Used for access authentication to foreman
type userAccess struct {
	username	string
	password	string
	url				string
}

type fRespDomain struct {
 Id		int			`json:"id"`
 Name	string	`json:"name"`
}
//This sets up the schema, the interface between the tf file and the plugin
func resourceServer() *schema.Resource {
	return &schema.Resource{
		Create: resourceServerCreate,
		Read:   resourceServerRead,
		Update: resourceServerUpdate,
		Delete: resourceServerDelete,

		Schema: map[string]*schema.Schema{
			"debug": &schema.Schema{
				Type:			schema.TypeBool,
				Optional:	true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"username": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"password": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"url": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"environment_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"ip": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"mac": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"architecture_id": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},
			"domain_id": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},
			"realm_id": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},
			"puppet_proxy_id": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},
			"puppetclass_ids": &schema.Schema{
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
				Set: func(v interface{}) int {
					return hashcode.String(v.(string))
				},
			},
			"operatingsystem_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"medium_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"ptable_id": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},
			"subnet_id": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},
			"compute_resource_id": &schema.Schema{
				Type:     schema.TypeInt, // if nil it assumes bare-metal build ,Add some validation logic later when you have time.
				Optional: true,
			},
			"root_pass": &schema.Schema{
				Type:     schema.TypeString, // required if host is managed and not inherited from hostgroup or default password.
				Optional: true,
			},
			"model_id": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},
			"hostgroup_id": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},
			"owner_id": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},
			"puppet_ca_proxy_id": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},
			"image_id": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},
			/*
			"host_parameters_attributes": &schema.Schema{
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
				Set: func(v interface{}) int {
					return hashcode.String(v.(string))
				},
			},
			*/
			"build": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
			},
			"enabled": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
			},
			"provision_method": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"managed": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
			},
			"progress_report_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"comment": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"capabilities": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"compute_profile_id": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},
			"interfaces_attributes": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"mac": &schema.Schema{
							Type:	schema.TypeString,
							Optional: true,
							ForceNew: false,
						},
						"ip": &schema.Schema{
							Type:	schema.TypeString,
							Optional: true,
							ForceNew: false,
						},
						"type": &schema.Schema{
							Type:	schema.TypeString,
							Optional: true,
							ForceNew: false,
						},
						"name": &schema.Schema{
							Type:	schema.TypeString,
							Optional: true,
							ForceNew: false,
						},
						"subnet_id": &schema.Schema{
							Type:	schema.TypeInt,
							Optional: true,
							ForceNew: false,
						},
						"domain_id": &schema.Schema{
							Type:	schema.TypeInt,
							Optional: true,
							ForceNew: false,
						},
						"identifier": &schema.Schema{
							Type:	schema.TypeString,
							Optional: true,
							ForceNew: false,
						},
						"managed": &schema.Schema{
							Type:	schema.TypeBool,
							Optional: true,
							ForceNew: false,
						},
						"primary": &schema.Schema{
							Type:	schema.TypeBool,
							Optional: true,
							ForceNew: false,
						},
						"provision": &schema.Schema{
							Type:	schema.TypeBool,
							Optional: true,
							ForceNew: false,
						},
						"username": &schema.Schema{
							Type:	schema.TypeString,
							Optional: true,
							ForceNew: false,
						},
						"password": &schema.Schema{
							Type:	schema.TypeString,
							Optional: true,
							ForceNew: false,
						},
						"provider": &schema.Schema{
							Type:	schema.TypeString,
							Optional: true,
							ForceNew: false,
						},
						"virtual": &schema.Schema{
							Type:	schema.TypeBool,
							Optional: true,
							ForceNew: false,
						},
						"tag": &schema.Schema{
							Type:	schema.TypeString,
							Optional: true,
							ForceNew: false,
						},
						"attached_to": &schema.Schema{
							Type:	schema.TypeString,
							Optional: true,
							ForceNew: false,
						},
						"mode": &schema.Schema{
							Type:	schema.TypeString,
							Optional: true,
							ForceNew: false,
						},
						"attached_devices": &schema.Schema{
							Type:	schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{Type: schema.TypeString},
							ForceNew: false,
						},
						"bond_options": &schema.Schema{
							Type:	schema.TypeString,
							Optional: true,
							ForceNew: false,
						},
						"compute_attributes": &schema.Schema{
							Type:     schema.TypeMap,
							Optional: true,
							ForceNew: false,
						},
					},
				},
			},
			"volumes_attributes": &schema.Schema{
				Type:		schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": &schema.Schema{
							Type:	schema.TypeString,
							Optional: true,
							ForceNew: false,
						},
						"size_gb": &schema.Schema{
							Type:	schema.TypeInt,
							Optional: true,
							ForceNew: false,
						},
						"_delete": &schema.Schema{
							Type:	schema.TypeBool,
							Optional: true,
							ForceNew: false,
						},
						"datastore": &schema.Schema{
							Type:	schema.TypeString,
							Optional: true,
							ForceNew: false,
						},
					},
				},
			},
			"compute_attributes": &schema.Schema{
				Type:     schema.TypeMap,
				Optional: true,
			},
			"host_parameters_attributes": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
							"roles" : &schema.Schema{
								Type: schema.TypeString,
								Optional: true,
								ForceNew: false,
							},
							"puppet" : &schema.Schema{
								Type: schema.TypeString,
								Optional: true,
								ForceNew: false,
							},
							"chef" : &schema.Schema{
								Type: schema.TypeString,
								Optional: true,
								ForceNew: false,
							},
							"JIRA_Ticket" : &schema.Schema{
								Type: schema.TypeString,
								Optional: true,
								ForceNew: false,
							},
						},
					},
			},
		},
	}
}

// Setup a function to make api calls
func httpClient(rType string, d *host, u *userAccess, apiSection string, debug bool, fqdn string) ([]byte, error ) {
  //setup local vars
  r := strings.ToUpper(rType)
  lUserAccess := u
  rHost := reqHost{}
	rHost.Lhost = *d

  b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(rHost)
  //build and make request
	client := &http.Client{}
	reqURL := ""
	//Need to account for different parts of the API
	switch apiSection{
		//If i want to work on hosts API
		case "hosts":
	  	switch r {
				case "POST","PUT":
					reqURL = fmt.Sprintf("%s/%s", lUserAccess.url, apiSection)
				case "GET","DELETE":
					reqURL = fmt.Sprintf("%s/%s/%s", lUserAccess.url, apiSection, fqdn)
				}
		//If I want to work on domains API
		case "domains":
			switch r {
			case "GET":
				reqURL = fmt.Sprintf("%s/%s/%d",lUserAccess.url, apiSection, rHost.Lhost.Domain_id)
			}
		}
	req, err := http.NewRequest(r,reqURL,b)
	if err != nil {
		panic(err)
	}
	//set basic auth if necessary
	if u.username != "" {
	req.SetBasicAuth(lUserAccess.username,lUserAccess.password)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json;version=2")
	req.Header.Add("Foreman_api_version", "2")
	req.Header.Add("Accept-Encoding", "gzip, deflate")
   //enable debugging data
	if debug {
		panic(req)
	}
  resp, err := client.Do(req)

	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	if content != nil {
		fmt.Println("%v",content)
	}

	return content, err
}

func getDomain(h *host, u *userAccess) string {
	dStruct := new(fRespDomain)
	resp, err := httpClient("GET", h, u, "domains", false,"")
	if resp != nil {
		fResp := fmt.Sprintf("The server responded with: %v",resp)
		print(fResp)
		if strings.Contains(string(resp),"error"){
			err = errors.New(string(resp))
		}
	}

	if err != nil {
		panic(err)
	}

	unerr := json.Unmarshal(resp,&dStruct)
	if unerr != nil {
		panic(unerr)
	}
	return dStruct.Name
}

func buildUserStruct(d *schema.ResourceData, meta interface{}) userAccess {
	u := userAccess{}
	/* populate u struct instance */
	if v, ok := d.GetOk("username"); ok {
		u.username = v.(string)
	}
	if v, ok := d.GetOk("password"); ok {
		u.password = v.(string)
	}
	if v, ok := d.GetOk("url"); ok {
		u.url	= v.(string)
	}
  return u
}

func buildHostStruct(d *schema.ResourceData, meta interface{}) host {
	h := host{}
	if v, ok := d.GetOk("name"); ok {
		h.Name = v.(string)
	}
	caprefix := fmt.Sprintf("compute_attributes")
	if v, ok := d.GetOk(caprefix+".cpus"); ok {
		h.Lcompute_attributes.Cpus = v.(string)
	}
	if v, ok := d.GetOk(caprefix+".start"); ok {
		h.Lcompute_attributes.Start = v.(string)
	}
	if v, ok := d.GetOk(caprefix+".cluster"); ok {
		h.Lcompute_attributes.Cluster = v.(string)
	}
	if v, ok := d.GetOk(caprefix+".memory_mb"); ok {
		h.Lcompute_attributes.Memory_mb = v.(string)
	}
	if v, ok := d.GetOk(caprefix+".guest_id"); ok {
		h.Lcompute_attributes.Guest_id = v.(string)
	}
/* build volumes_attributes now, this needs to be mapped so I build structs and then add them to the mapped value */
vaCount := d.Get("volumes_attributes.#").(int)
if vaCount >0 {
	h.Lcompute_attributes.Lvolumes_attributes = make(map[string]volumes_attributes)
for i := 0; i<vaCount; i++ {
	// setup iterator market and instantiate local struct to append to maps
	iStr:=fmt.Sprintf("%d",i)
	lStruct := volumes_attributes{}
	//setup lStruct values
	vaprefix := fmt.Sprintf("volumes_attributes.%d",i)
	if v, ok := d.GetOk(vaprefix+".name"); ok {
		lStruct.Name = v.(string)
	}
	if v, ok := d.GetOk(vaprefix+".size_gb"); ok {
		lStruct.Size_gb = v.(int)
	}
	if v, ok := d.GetOk(vaprefix+"._delete"); ok {
		lStruct._delete = v.(string)
	}
	if v, ok := d.GetOk(vaprefix+".datastore"); ok {
		lStruct.Datastore = v.(string)
	}
	//add in lStruct to the main host struct
	h.Lcompute_attributes.Lvolumes_attributes[iStr] = lStruct
}
}
/* build interfaces_attributes now */
iaCount := d.Get("interfaces_attributes.#").(int)
if iaCount >0 {
for i := 0; i<iaCount; i++ {
	h.Linterfaces_attributes = append(h.Linterfaces_attributes,interfaces_attributes{})
	prefix := fmt.Sprintf("interfaces_attributes.%d",i)
	if v, ok := d.GetOk(prefix+".primary"); ok {
		h.Linterfaces_attributes[i].Primary = v.(bool)
	}
	//Adding some logic to auto populate primary nic because of API deferral
	if v, ok := d.GetOk(prefix+".mac"); ok {
		h.Linterfaces_attributes[i].Mac = v.(string)
	} /* else if v, ok := d.GetOk("mac"); ok && h.Linterfaces_attributes[i].Primary {
		h.Linterfaces_attributes[i].Mac = v.(string)
	} */
	if v, ok := d.GetOk(prefix+".ip"); ok {
		h.Linterfaces_attributes[i].Ip = v.(string)
	} /* else if v, ok := d.GetOk("ip"); ok && h.Linterfaces_attributes[i].Primary {
		h.Linterfaces_attributes[i].Ip = v.(string)
	} */
	if v, ok := d.GetOk(prefix+".type"); ok {
		h.Linterfaces_attributes[i].Type = v.(string)
	}
	if v, ok := d.GetOk(prefix+".name"); ok {
		h.Linterfaces_attributes[i].Name = v.(string)
	} /* else if v, ok := d.GetOk("name"); ok && h.Linterfaces_attributes[i].Primary {
		h.Linterfaces_attributes[i].Name = v.(string)
	} */
	if v, ok := d.GetOk(prefix+".subnet_id"); ok {
		h.Linterfaces_attributes[i].Subnet_id = v.(int)
	} /* else if v, ok := d.GetOk("subnet_id"); ok && h.Linterfaces_attributes[i].Primary {
		h.Linterfaces_attributes[i].Subnet_id = v.(int)
	} */
	if v, ok := d.GetOk(prefix+".domain_id"); ok {
		h.Linterfaces_attributes[i].Domain_id = v.(int)
	} /*else if v, ok := d.GetOk("domain_id"); ok && h.Linterfaces_attributes[i].Primary {
		h.Linterfaces_attributes[i].Domain_id = v.(int)
	} */
	if v, ok := d.GetOk(prefix+".identifier"); ok {
		h.Linterfaces_attributes[i].Identifier = v.(string)
	}
	if v, ok := d.GetOk(prefix+".managed"); ok {
		h.Linterfaces_attributes[i].Managed = v.(bool)
	}
	if v, ok := d.GetOk(prefix+".provision"); ok {
		h.Linterfaces_attributes[i].Provision = v.(bool)
	}
	if v, ok := d.GetOk(prefix+".username"); ok {
		h.Linterfaces_attributes[i].Username = v.(string)
	}
	if v, ok := d.GetOk(prefix+".password"); ok {
		h.Linterfaces_attributes[i].Password = v.(string)
	}
	if v, ok := d.GetOk(prefix+".provider"); ok {
		h.Linterfaces_attributes[i].Provider = v.(string)
	}
	if v, ok := d.GetOk(prefix+".virtual"); ok{
		h.Linterfaces_attributes[i].Virtual = v.(bool)
	}
	if v, ok := d.GetOk(prefix+".tag"); ok {
		h.Linterfaces_attributes[i].Tag = v.(string)
	}
	if v, ok := d.GetOk(prefix+".attached_to"); ok {
		h.Linterfaces_attributes[i].Attached_to = v.(string)
	}
	if v, ok := d.GetOk(prefix+".mode"); ok {
		h.Linterfaces_attributes[i].Mode = v.(string)
	}
	if v, ok := d.GetOk(prefix+".attached_devices"); ok {
		h.Linterfaces_attributes[i].Attached_devices = v.([]string)
	}
	if v, ok := d.GetOk(prefix+".bond_options"); ok {
		h.Linterfaces_attributes[i].Bond_options = v.(string)
	}
	ifcaprefix := fmt.Sprintf("%s.compute_attributes",prefix)
	if v, ok := d.GetOk(ifcaprefix+".network"); ok {
		h.Linterfaces_attributes[i].Lcompute_attributes.Network = v.(string)
	}
	if v, ok := d.GetOk(ifcaprefix+".type"); ok {
		h.Linterfaces_attributes[i].Lcompute_attributes.Type = v.(string)
	}
	}
}


/* populate host_parameters_attributes now */
hpaCount := d.Get("host_parameters_attributes.#").(int)
if hpaCount > 0 {
	h.Lhost_parameters_attributes = []params_archetype{}
for i := 0; i<hpaCount; i++ {
	intCnt := 0
	//append(h.Lhost_parameters_attributes,host_parameters_attributes{})
	prefix := fmt.Sprintf("host_parameters_attributes.%d",i)
	if v, ok := d.GetOk(prefix+".roles"); ok {
		h.Lhost_parameters_attributes = append(h.Lhost_parameters_attributes.Parameters,params_archetype{})
		h.Lhost_parameters_attributes.Parameters[i].Name="roles"
		h.Lhost_parameters_attributes.Parameters[i].Value=v.(string)
		//h.Lhost_parameters_attributes[i] = params_archetype{Name: "roles",Value: v.(string)}
	}
	if v, ok := d.GetOk(prefix+".puppet"); ok {
		//h.Lhost_parameters_attributes[i] = params_archetype{Name: "puppet",Value: v.(string)}
		h.Lhost_parameters_attributes = append(h.Lhost_parameters_attributes.Parameters,params_archetype{})
		h.Lhost_parameters_attributes.Parameters[i].Name="puppet"
		h.Lhost_parameters_attributes.Parameters[i].Value=v.(string)
	}
	if v, ok := d.GetOk(prefix+".chef"); ok {
		//h.Lhost_parameters_attributes[i] = params_archetype{Name: "chef",Value: v.(string)}
		h.Lhost_parameters_attributes = append(h.Lhost_parameters_attributes.Parameters,params_archetype{})
		h.Lhost_parameters_attributes.Parameters[i].Name="chef"
		h.Lhost_parameters_attributes.Parameters[i].Value=v.(string)
	}
	if v, ok := d.GetOk(prefix+".JIRA_Ticket"); ok {
		//h.Lhost_parameters_attributes[i] = params_archetype{Name: "JIRA_Ticket",Value: v.(string)}
		h.Lhost_parameters_attributes = append(h.Lhost_parameters_attributes.Parameters,params_archetype{})
		h.Lhost_parameters_attributes.Parameters[i].Name="JIRA_Ticket"
		h.Lhost_parameters_attributes.Parameters[i].Value=v.(string)
	}
}
}

/* populate h struct instance for regular level data */
	if v, ok := d.GetOk("environment_id"); ok {
		h.Environment_id = v.(string)
	}
	if v,ok := d.GetOk("ip"); ok{
		h.Ip = v.(string)
	}
	if v,ok := d.GetOk("mac"); ok{
		h.Mac = v.(string)
	}
	if v,ok := d.GetOk("architecture_id"); ok{
		h.Architecture_id = v.(int)
	}
	if v,ok := d.GetOk("domain_id"); ok{
		h.Domain_id = v.(int)
	}
	if v,ok := d.GetOk("realm_id"); ok{
		h.Realm_id = v.(int)
	}
	if v,ok := d.GetOk("puppet_proxy_id"); ok{
		h.Puppet_proxy_id = v.(int)
	}
	if v,ok := d.GetOk("puppet_class_ids"); ok{
		h.Puppetclass_ids = v.([]int)
	}
	if v,ok := d.GetOk("operatingsystem_id"); ok{
		h.Operatingsystem_id = v.(string)
	}
	if v,ok := d.GetOk("medium_id"); ok{
		h.Medium_id = v.(string)
	}
	if v,ok := d.GetOk("ptable_id"); ok{
		h.Ptable_id = v.(int)
	}
	if v,ok := d.GetOk("subnet_id"); ok{
		h.Subnet_id = v.(int)
	}
	if v,ok := d.GetOk("compute_resource_id"); ok{
		h.Compute_resource_id = v.(int)
	}
	if v,ok := d.GetOk("root_pass"); ok{
		h.Root_pass = v.(string)
	}
	if v,ok := d.GetOk("model_id"); ok{
		h.Model_id = v.(int)
	}
	if v,ok := d.GetOk("hostgroup_id"); ok{
		h.Hostgroup_id = v.(int)
	}
	if v,ok := d.GetOk("owner_id"); ok{
		h.Owner_id = v.(int)
	}
	if v,ok := d.GetOk("owner_type"); ok{
		if v.(string) == "User" || v.(string) == "Usergroup" {
			h.Owner_type = v.(string)
		}
	}
	if v,ok := d.GetOk("puppet_ca_proxy_id"); ok{
		h.Puppet_ca_proxy_id = v.(int)
	}
	if v,ok := d.GetOk("image_id"); ok{
		h.Image_id = v.(int)
	}
	if v,ok := d.GetOk("build"); ok{
		h.Build = v.(bool)
	}
	if v,ok := d.GetOk("enabled"); ok{
		h.Enabled = v.(bool)
	}
	if v,ok := d.GetOk("provision_method"); ok{
		h.Provision_method = v.(string)
	}
	if v,ok := d.GetOk("managed"); ok{
		h.Managed = v.(bool)
	}
	if v,ok := d.GetOk("progress_report_id"); ok{
		h.Progress_report_id = v.(string)
	}
	if v,ok := d.GetOk("comment"); ok{
		h.Comment = v.(string)
	}
	if v,ok := d.GetOk("capabilities"); ok{
		h.Capabilities = v.(string)
	}
	if v,ok := d.GetOk("compute_profile_id"); ok{
		h.Compute_profile_id = v.(int)
	}
	return h
}

func resourceServerCreate(d *schema.ResourceData, meta interface{}) error {
  h := buildHostStruct(d,meta)
	u := buildUserStruct(d,meta)

	/* check debug flag */
	debug := false
	if v, ok := d.GetOk("debug"); ok {
		debug = v.(bool)
	}

	resp, err := httpClient("POST", &h, &u, "hosts", debug,"")
	if resp != nil {
		fResp := fmt.Sprintf("The server responded with: %v",resp)
		print(fResp)
		if strings.Contains(string(resp),"error"){
			err = errors.New(string(resp))
		}
	}

	if err != nil {
		return err
	}
	d.SetId(d.Get("name").(string))
	return nil
}

func resourceServerRead(d *schema.ResourceData, m interface{}) error {
	h := buildHostStruct(d,m)
	u := buildUserStruct(d,m)
  dom := getDomain(&h,&u)
	fqdn := fmt.Sprintf("%s.%s",h.Name,dom)

	resp, err := httpClient("GET", &h, &u, "hosts", false,fqdn)
	if resp != nil {
		fResp := fmt.Sprintf("The server responded with: %v",resp)
		print(fResp)
		if strings.Contains(string(resp),"error"){
			err = errors.New(string(resp))
		}
	}

	if err != nil {
		return err
	}
	d.SetId(d.Get("name").(string))
	d.Set("ip",h.Ip)
	return nil
}

func resourceServerUpdate(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceServerDelete(d *schema.ResourceData, m interface{}) error {
	h := buildHostStruct(d,m)
	u := buildUserStruct(d,m)
	dom := getDomain(&h,&u)
	fqdn := fmt.Sprintf("%s.%s",h.Name,dom)

	resp, err := httpClient("DELETE", &h, &u, "hosts", false,fqdn)
	if resp != nil {
		fResp := fmt.Sprintf("The server responded with: %v",resp)
		print(fResp)
		if strings.Contains(string(resp),"error"){
			err = errors.New(string(resp))
		}
	}

	if err != nil {
		return err
	}

	return nil
}
