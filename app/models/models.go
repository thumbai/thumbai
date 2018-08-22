package models

// PackageInfo holds single Go package path belongs to a host.
type PackageInfo struct {
	Path string `json:"path,omitempty"`
	Repo string `json:"repo,omitempty"`
	VCS  string `json:"vcs,omitempty"`
	Src  string `json:"-"`
}

// ProxyInfo holds single Reverse Proxy server info.
type ProxyInfo struct {
	URL       string      `json:"url,omitempty"`
	Redirects []*Redirect `json:"redirects,omitempty"`
	ReqHdr    *Hdr        `json:"request_header,omitempty"`
	ResHdr    *Hdr        `json:"response_header,omitempty"`
}

// Redirect holds single redirect for proxy server.
type Redirect struct {
	Match  string
	Target string
	Code   int
}

// Hdr struct holds the request needs to be added or removed.
type Hdr struct {
	Add    map[string]string `json:"add,omitempty"`
	Remove []string          `json:"remove,omitempty"`
}
