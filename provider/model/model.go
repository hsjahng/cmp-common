package model

import "encoding/json"

type Scope struct {
	NodeType string `json:"nodeType"` // instance, host, vm, datastore
}

type Platform struct {
	ProviderId string `json:"providerId"`
	ObjectType string `json:"objectType"` // vmware, aws...
	DataType   string `json:"dataType"`   // meta, metric
	Name       string `json:"name"`       // platform 이름
}

type ScopeMeta struct {
	Scope Scope             `json:"scope"`
	Data  []json.RawMessage `json:"metas"`
}

type CommonMetaModel struct {
	Resource  Platform  `json:"resource"`
	ScopeMeta ScopeMeta `json:"scopeMeta"`
}

type ParsedMetaModel struct {
	ObjectType string      // "aws", "vmware" 등
	NodeType   string      // "instance", "host", "vm", "datastore"
	ProviderId string      // AWS, VMware Provider ID
	Data       interface{} // 파싱된 데이터 (AWSInstance, VMWareVM 등)
}

type ProviderType int

const (
	OPENSTACK ProviderType = 1
	K8S       ProviderType = 2
	NCP       ProviderType = 3
	AWS       ProviderType = 4
	VSPHERE   ProviderType = 5
	NETAPP    ProviderType = 6
)

func (p ProviderType) Int() int {
	return int(p)
}
