package service

type Node struct {
	ID          string   `json:"id,omitempty"`
	Title       string   `json:"title,omitempty"`
	Handle      string   `json:"handle,omitempty"`
	Vendor      string   `json:"vendor,omitempty"`
	ProductType string   `json:"producType,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Namespace   string   `json:"namespace,omitempty"`
	Key         string   `json:"key,omitempty"`
	Value       string   `json:"value,omitempty"`
	ParentID    string   `json:"__parentId,omitempty"`
}

type Product struct {
	ID          string      `json:"id,omitempty"`
	Title       string      `json:"title,omitempty"`
	Handle      string      `json:"handle,omitempty"`
	Vendor      string      `json:"vendor,omitempty"`
	ProductType string      `json:"producType,omitempty"`
	Tags        []string    `json:"tags,omitempty"`
	Metafields  []Metafield `json:"metafields,omitempty"`
}

type Metafield struct {
	Namespace string `json:"namespace,omitempty"`
	Key       string `json:"key,omitempty"`
	Value     string `json:"value,omitempty"`
	ParentID  string `json:"__parentId,omitempty"`
}
