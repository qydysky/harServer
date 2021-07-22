package main

import (
	"time"
	"net/url"
)

type Har struct {
	Log struct {
		Entries []struct {
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
		Encoding    string `json:""encoding"`
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
	Url string
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