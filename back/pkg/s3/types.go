package s3

type NatsHasRequest struct {
	Key string `json:"key,omitempty"`
}

type NatsHasResponse struct {
	Has bool `json:"has,omitempty"`
}

type NatsGetRequest struct {
	Key string `json:"key,omitempty"`
}

type NatsGetResponse struct {
	Error string `json:"error,omitempty"`
	Url   string `json:"url,omitempty"`
	Key   string `json:"key,omitempty"`
}

type NatsPutRequest struct {
	Folder     string `json:"folder,omitempty"`
	Filename   string `json:"filename,omitempty"`
	Key        string `json:"key,omitempty"`
	NeedGetUrl bool   `json:"need_get_url,omitempty"`
	KeepName   bool   `json:"keep_name,omitempty"`
}

type NatsPutResponse struct {
	Error  string `json:"error,omitempty"`
	Url    string `json:"url,omitempty"`
	GetUrl string `json:"get_url,omitempty"`
	Key    string `json:"key,omitempty"`
}

type NatsGetMultipleRequest struct {
	Keys []string `json:"keys,omitempty"`
}
type NatsGetMultipleResponse struct {
	Error string   `json:"error,omitempty"`
	Urls  []string `json:"urls,omitempty"`
	Keys  []string `json:"keys,omitempty"`
}

type NatsPutMultipleRequest struct {
	Folder    string   `json:"folder,omitempty"`
	Filenames []string `json:"filenames,omitempty"`
	KeepNames bool     `json:"keep_names,omitempty"`
}

type NatsPutMultipleResponse struct {
	Error string   `json:"error,omitempty"`
	Urls  []string `json:"urls,omitempty"`
	Keys  []string `json:"keys,omitempty"`
}

type NatsDeleteRequest struct {
	Key string `json:"key,omitempty"`
}

type NatsDeleteResponse struct {
	Error string `json:"error,omitempty"`
	Ok    bool   `json:"ok,omitempty"`
}
