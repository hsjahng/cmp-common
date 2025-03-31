package kafka

import "encoding/json"

type ResourceModel struct {
	Resource  Resource  `json:"resource"`
	ScopeData ScopeData `json:"scopeData"`
}

type Resource struct {
	ProviderId string `json:"providerId"`
	ObjectType string `json:"objectType"` // vmware, aws...
	DataType   string `json:"dataType"`   // meta, metric
	Name       string `json:"name"`       // 미사용
}

type ScopeData struct {
	Scope Scope             `json:"scope"`
	Data  []json.RawMessage `json:"datas"`
}

type Scope struct {
	NodeType string `json:"nodeType"` // instance, host, vm, datastore
}

type KafkaParsedModel struct {
	ObjectType string      // "aws", "vmware" 등
	NodeType   string      // "instance", "host", "vm", "datastore"
	ProviderId string      // AWS, VMware Provider ID
	Data       interface{} // 파싱된 데이터 (AWSInstance, VMWareVM 등)
}
