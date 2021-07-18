package main

import (
	"fmt"
	"log"
	"time"
	"flag"
	"bytes"
	"strings"
	"context"
	"net/url"
	"net/http"
	"encoding/json"

	part "github.com/qydysky/part"
	partWeb "github.com/qydysky/part/web"
	compress "github.com/qydysky/part/compress"
)

type RequestResponse struct {
	req harRequest
	res harResponse
}

func main(){

	var listenAddr = flag.String("l", "127.0.0.1:2222", "-l listen address(eg. 127.0.0.1:2222)(default: 127.0.0.1:2222)")
	var harFiles = flag.String("f", "test.har", "-f har files(eg. test.har,test1.har)(default: test.har)")
	flag.Parse()
	log.Println(`使用监听地址`, *listenAddr)
	log.Println(`使用hars文件`, *harFiles)

	var (
		urlMap = make(map[string]([]RequestResponse))
		removeStr = []string{"integrity=", "https:", "http:"}
		replaceStr = []string{}
	)
	for _,harFile := range strings.Split(*harFiles, ",") {
		harByte := ReadHar(harFile)
		if len(harByte) == 0 {
			log.Println(`读取har错误`, harFile)
			continue
		} else {
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
			} else if v.Response.Status != 0 {
				replaceStr = append(replaceStr, u.Scheme+`://`+u.Host)
				replaceStr = append(replaceStr, u.Host)
				urlMap[u.Path] = append(urlMap[u.Path], RequestResponse{
					req: harRequest{
						Id: fmt.Sprintf("F:%s;I:%s;", harFile, v.Connection),
						Method: v.Request.Method,
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
		Addr: *listenAddr,
		WriteTimeout:  time.Second * time.Duration(10),
	})

	var webServerAddr string
	if addr := strings.Split(*listenAddr, ":"); addr[0] == "0.0.0.0" {
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
				
				maxV := match(r, vs)

				if maxV.req.Id != `` {
					log.Println(`[ok] req:`, path, maxV.req.Id)

					//COS
					w.Header().Set("Access-Control-Allow-Origin", "*")//允许访问所有域
					w.Header().Add("Access-Control-Allow-Headers", "*")//header的类型
	
					//Headers
					var contentEncoding string
					for _,header :=range maxV.res.Headers {
						if header.Name == "content-encoding" || header.Value == "gzip" {
							contentEncoding = header.Value
						}
						if header.Name == "content-security-policy" ||
						   header.Name == "content-length" {continue}
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
	
					//statusCode
					w.WriteHeader(maxV.res.Status)
	
					//Content
					var responseData = []byte(maxV.res.Content.Text)
					if strings.Contains(maxV.res.Content.MimeType, `text/html`) {
						for _,v :=range replaceStr {
							responseData = bytes.ReplaceAll(responseData, []byte(v), []byte(webServerAddr))
						}
						for _,v :=range removeStr {
							responseData = bytes.ReplaceAll(responseData, []byte(v), []byte(""))
						}
					}
					if contentEncoding == "gzip" {
						var e error
						responseData,e = compress.InGzip(responseData, -1)
						if e != nil {
							log.Println(e)
						}
					}
					w.Write(responseData)
					return
				}
			} else {
				log.Println(`[no] `, path)
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

func ReadHar(src string) []byte {
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
			if harRR.req.PostData[k] == v {
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
