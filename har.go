package main

import (
	"time"
	"net/url"
)

type Har struct {
	Log struct {
		Version string `json:"version"`
		Creator struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"creator"`
		Pages   []interface{} `json:"pages"`
		Entries []struct {
			Initiator struct {
				Type  string `json:"type"`
				Stack struct {
					CallFrames []struct {
						FunctionName string `json:"functionName"`
						ScriptID     string `json:"scriptId"`
						URL          string `json:"url"`
						LineNumber   int    `json:"lineNumber"`
						ColumnNumber int    `json:"columnNumber"`
					} `json:"callFrames"`
					Parent struct {
						Description string `json:"description"`
						CallFrames  []struct {
							FunctionName string `json:"functionName"`
							ScriptID     string `json:"scriptId"`
							URL          string `json:"url"`
							LineNumber   int    `json:"lineNumber"`
							ColumnNumber int    `json:"columnNumber"`
						} `json:"callFrames"`
						Parent struct {
							Description string `json:"description"`
							CallFrames  []struct {
								FunctionName string `json:"functionName"`
								ScriptID     string `json:"scriptId"`
								URL          string `json:"url"`
								LineNumber   int    `json:"lineNumber"`
								ColumnNumber int    `json:"columnNumber"`
							} `json:"callFrames"`
						} `json:"parent"`
					} `json:"parent"`
				} `json:"stack"`
			} `json:"_initiator"`
			Priority     string `json:"_priority"`
			ResourceType string `json:"_resourceType"`
			Cache        struct {
			} `json:"cache"`
			Connection string `json:"connection"`
			Request    struct {
				Method      string `json:"method"`
				URL         string `json:"url"`
				HTTPVersion string `json:"httpVersion"`
				Headers     []struct {
					Name  string `json:"name"`
					Value string `json:"value"`
				} `json:"headers"`
				QueryString []interface{} `json:"queryString"`
				Cookies     []struct {
					Name     string    `json:"name"`
					Value    string    `json:"value"`
					Path     string    `json:"path"`
					Domain   string    `json:"domain"`
					Expires  time.Time `json:"expires"`
					HTTPOnly bool      `json:"httpOnly"`
					Secure   bool      `json:"secure"`
				} `json:"cookies"`
				HeadersSize int `json:"headersSize"`
				BodySize    int `json:"bodySize"`
				PostData    struct {
					MimeType string `json:"mimeType"`
					Text     string `json:"text"`
				} `json:"postData"`
			} `json:"request"`
			Response harResponse `json:"response"`
			ServerIPAddress string    `json:"serverIPAddress"`
			StartedDateTime time.Time `json:"startedDateTime"`
			Time            float64   `json:"time"`
			Timings         struct {
				Blocked         float64 `json:"blocked"`
				DNS             float64 `json:"dns"`
				Ssl             int     `json:"ssl"`
				Connect         float64 `json:"connect"`
				Send            float64 `json:"send"`
				Wait            float64 `json:"wait"`
				Receive         float64 `json:"receive"`
				BlockedQueueing float64 `json:"_blocked_queueing"`
			} `json:"timings"`
		} `json:"entries"`
	} `json:"log"`
}

type harResponse struct {
	Status      int    `json:"status"`
	StatusText  string `json:"statusText"`
	HTTPVersion string `json:"httpVersion"`
	Headers     []struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	} `json:"headers"`
	Cookies []harCookie `json:"cookies"`
	Content struct {
		Size        int    `json:"size"`
		MimeType    string `json:"mimeType"`
		Compression int    `json:"compression"`
		Text        string `json:"text"`
	} `json:"content"`
	RedirectURL  string      `json:"redirectURL"`
	HeadersSize  int         `json:"headersSize"`
	BodySize     int         `json:"bodySize"`
	TransferSize int         `json:"_transferSize"`
	Error        interface{} `json:"_error"`
}

type harRequest struct {
	Id string
	Method string
	Path string
	Query url.Values
	Fragment string
	PostData string
}

type harCookie struct {
	Name     string    `json:"name"`
	Value    string    `json:"value"`
	Domain   string    `json:"domain"`
	Expires  time.Time `json:"expires"`
	Httponly bool      `json:"httpOnly"`
	Secure   bool      `json:"secure"`
}