## HAR webServer
### Introduction
set up a web server from your [har file](https://en.wikipedia.org/wiki/HAR_(file_format))

### attention
due to the [bug](https://stackoverflow.com/questions/38924798/chrome-dev-tools-fails-to-show-response-even-the-content-returned-has-header-con), i recommond to use firefox to get har file.

### demo
- clone this Repository
- go run . -c main.json
- open http://127.0.0.1:2222/gitbook-documentation/content/index.html

### help
run `go run . -h` to get more help

### about default config
```json
{
    "ListenAddr": "127.0.0.1:2222",//default server addr, can rewrite by option -l
    "HarFiles": [//default har files, can ADD by option -f
        "test.har"
    ],
    "ResponseHeader": [//server will add the following headers to any request
        {
            "Name": "Access-Control-Allow-Origin",
            "Value": "*"
        },
        {
            "Name": "Access-Control-Allow-Methods",
            "Value": "POST, GET, PUT, OPTIONS, DELETE"
        },
        {
            "Name": "Access-Control-Allow-Headers",
            "Value": "*,content-type,authorization"
        },
		{
			"Name": "Content-Encoding",// will encode to gzip, Only gzip support now
			"Value": "gzip"
		}
    ],
    "IgnoreHeader": [//server will ignore the follow headers in har respone
        "content-security-policy",
        "content-length",
		"Content-Encoding"
    ],
    "RemoveString": [//remove the following strings when respone the html page
        "https:",
        "http:",
        "integrity="
    ],
	"Log": {
		"success": true,// log success requst
		"fail": true// log fail requst
	}
}
```