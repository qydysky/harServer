package main

import (
	"fmt"
	"log"
	"time"
	"flag"
	"bytes"
	"unicode"
	"strings"
	"context"
	"net/url"
	"net/http"
	"encoding/json"
	"encoding/base64"

	part "github.com/qydysky/part"
	partWeb "github.com/qydysky/part/web"
	reqf "github.com/qydysky/part/reqf"
	compress "github.com/qydysky/part/compress"
)

type RequestResponse struct {
	req harRequest
	res harResponse
}

type Config struct {
	ListenAddr     string   `json:"ListenAddr"`
	HarFiles       []string `json:"HarFiles"`
	ResponseHeader []struct {
		Name  string `json:"Name"`
		Value string `json:"Value"`
	} `json:"ResponseHeader"`
	IgnoreHeader []string `json:"IgnoreHeader"`
	RemoveString []string `json:"RemoveString"`
	Log          struct {
		Success bool `json:"success"`
		Fail    bool `json:"fail"`
	} `json:"Log"`

	ignoreHeaderMap map[string]struct{} // just make it fast
	replaceStr []string // Contain host strings of all files, will be replace to localhost
}

func main(){

	//全局配置
	var globalConfig Config

	//加载配置
	{
		var (
			listenAddr = flag.String("l", "", "-l listen address(eg. 127.0.0.1:2222)")
			harFiles = flag.String("f", "", "-f har files(eg. test.har,test1.har)")
			config = flag.String("c", "", "-c config file(eg. main.json)")
		)
		flag.Parse()

		if configByte := Read(*config);len(configByte) != 0 {
			if e := json.Unmarshal(configByte, &globalConfig);e != nil {
				log.Println(`配置文件错误`, e)
			} else {
				log.Println(`已加载配置文件`, *config)
				globalConfig.ignoreHeaderMap = make(map[string]struct{})
				for _,v :=range globalConfig.IgnoreHeader {
					globalConfig.ignoreHeaderMap[strings.ToLower(v)] = struct{}{}
				}
			}
		}
		if *listenAddr != `` {
			log.Println(`使用监听地址`, *listenAddr)
			globalConfig.ListenAddr = *listenAddr
		}
		if *harFiles != `` {
			log.Println(`使用hars文件`, *harFiles)
			globalConfig.HarFiles = append(globalConfig.HarFiles, strings.Split(*harFiles, ",")...)
		}
		if globalConfig.ListenAddr == `` || len(globalConfig.HarFiles) == 0 {
			log.Println(`未指定监听地址或hars文件`)
			return
		}
	}
	
	//url->respone request map
	var urlMap = make(map[string]([]RequestResponse))

	//加载map
	for _,harFile := range globalConfig.HarFiles {
		harByte := Read(harFile)
		if len(harByte) == 0 {
			log.Println(`读取har错误`, harFile)
			continue
		} else {
			t := make([]byte, utf8.RuneCountInString(harByte))
			i := 0
			for _, r := range harByte {
				t[i] = byte(r)
				i++
			}
			log.Println(`已读取文件`, harFile)
		}
	
		var harJson Har
		if e := json.Unmarshal(harByte, &harJson);e != nil {
			log.Println(`读取har错误`, harFile, e)
			continue
		}
	
		for _,v :=range harJson.Log.Entries {
			if u,e := url.Parse(v.Request.URL);e != nil {
				log.Println("接口处理错误", e)
				continue
			} else if v.Response.Content.Size != 0 && v.Response.Content.Text == `` {
				continue
			} else {
				globalConfig.replaceStr = append(globalConfig.replaceStr, []string{
					u.Scheme+`://`+u.Host,
					`//`+u.Host,
					u.Host,
				}...)
				globalConfig.replaceStr = relist(globalConfig.replaceStr)
				urlMap[u.Path] = append(urlMap[u.Path], RequestResponse{
					req: harRequest{
						Id: fmt.Sprintf("=> %s", harFile),
						Method: v.Request.Method,
						Url: u.String(),
						Path: u.Path,
						Query: u.Query(),
						Fragment: u.EscapedFragment(),
						PostData: v.Request.PostData.Text,
					},
					res: v.Response,
				})
			}
		}
	}

	log.Println(`已加载路径`, len(urlMap), `条`)

	s := partWeb.New(&http.Server{
		Addr: globalConfig.ListenAddr,
		WriteTimeout:  time.Second * time.Duration(10),
	})

	var webServerAddr string
	if addr := strings.Split(globalConfig.ListenAddr, ":"); addr[0] == "0.0.0.0" {
		webServerAddr = `http://`+part.Sys().GetIntranetIp()+`:`+addr[1]
	} else {
		webServerAddr = `http://`+s.Server.Addr
	}
	log.Println(`harServer启动于`, webServerAddr)

	s.Handle(map[string]func(http.ResponseWriter,*http.Request){
		`/`:func(w http.ResponseWriter,r *http.Request){
			defer r.Body.Close()
			var path string = r.URL.Path
			
			if vs,ok := urlMap[path];ok {
				
				var maxV RequestResponse
				if len(vs) != 1 {
					maxV = match(r, vs)
				} else {
					maxV = vs[0]
				}

				if maxV.req.Id != `` {
					if globalConfig.Log.Success {log.Println(`[✔]`, path, maxV.req.Id)}

					//Headers
					for _,header :=range maxV.res.Headers {
						if _,ok := globalConfig.ignoreHeaderMap[strings.ToLower(header.Name)];ok {continue}
						w.Header().Set(header.Name, header.Value)
					}
	
					//Cookies
					for _,cookie :=range maxV.res.Cookies {
						w.Header().Set("Set-Cookie", (&http.Cookie{
							Name: cookie.Name,
							Value: cookie.Value,
							Domain: cookie.Domain,
							Expires: time.Now().Add(time.Hour),
							HttpOnly: cookie.Httponly,
							Secure: cookie.Secure,
						}).String())
					}

					//custom headers
					for _,v :=range globalConfig.ResponseHeader {
						w.Header().Set(v.Name, v.Value)
					}
						
					//statusCode
					if maxV.res.Status != 0 {
						w.WriteHeader(maxV.res.Status)
					}
	
					var responseData = []byte(maxV.res.Content.Text)
					//Content base64 decode
					if maxV.res.Content.Encoding == `base64` {
						if _,e := base64.StdEncoding.Decode(responseData, responseData);e != nil {
							log.Println(e)
						}
					}

					//Content
					if strings.Contains(maxV.res.Content.MimeType, `text/html`) {
						if globalConfig.Log.Success {log.Println(`[✔]`, maxV.req.Url)}
						r := reqf.New()
						if e := r.Reqf(reqf.Rval{
							Url: maxV.req.Url,
							Timeout: 5000,
						});e != nil {
							log.Println(e)
						} else {
							responseData = r.Respon
						}
						
						for _,v :=range globalConfig.replaceStr {
							responseData = bytes.ReplaceAll(responseData, []byte(v), []byte(webServerAddr))
						}
						for _,v :=range globalConfig.RemoveString {
							responseData = bytes.ReplaceAll(responseData, []byte(v), []byte(""))
						}
					}

					if w.Header().Get("Content-Encoding") == "gzip" {
						var e error
						responseData,e = compress.InGzip(responseData, -1)
						if e != nil {
							log.Println(e)
						}
					}
					w.Write(responseData)
					return
				}
			} else if globalConfig.Log.Fail {
				log.Println(`[✖]`, path)
			}

			if path == `/` {path = `index.html`}
			http.ServeFile(w, r, path)
		},
		`/exit`:func(w http.ResponseWriter,r *http.Request){
			s.Server.Shutdown(context.Background())
		},
	})

	log.Println(`使用ctrl+c退出服务`)

	for {
		time.Sleep(time.Second)
	}
}

func Read(src string) []byte {
	f := part.File()
	content := f.FileWR(part.Filel{
		File:src,
	})
	return []byte(content)
}

func match(req *http.Request,harRRs []RequestResponse) (maxV RequestResponse) {

	var reqPostData []byte
	if req.Body != nil {
		var PostData = make([]byte, 128)
		if n,e := req.Body.Read(PostData);e == nil || e.Error() == "EOF" {
			reqPostData = PostData[:n]
		}
	}

	var maxDiff int
	for _,harRR :=range harRRs {
		diff := 0

		if req.URL.Path != harRR.req.Path || req.Method != harRR.req.Method {
			continue
		} else {
			diff += 1
		}
	
		for k,v :=range req.URL.Query() {
			querys,ok := harRR.req.Query[k]
			if !ok {continue}
			if querys[0] == v[0] {diff += 1}
		}

		for k,v :=range reqPostData {
			if len(harRR.req.PostData) <= k {
				break
			} else if harRR.req.PostData[k] == v {
				diff += 1
			} else {
				break
			}
		}

		if maxDiff < diff {
			maxDiff = diff
			maxV = harRR
		}
	}
	return
}

func relist(s []string) (st []string) {
	for k,v :=range s {
		for sk,sv :=range st {
			if len(v) > len(sv) {
				st = append(st[:sk], append([]string{v}, st[sk:]...)...)
				break
			}
		}
		if len(st) != k+1 {
			st = append(st, v)
		}
	}

	return
} 