/*
 */

package apollo

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/pkg/errors"
	"net"
	"net/http"
	url2 "net/url"
	"time"
)

var (
	defalultClient *http.Client
	//defaultMaxConnsPerHost defines the maximum number of concurrent connections
	defaultMaxConnsPerHost = 8
	//defaultTimeoutBySecond defines the default timeout for http connections
	defaultTimeoutBySecond = 2 * time.Second
	//defaultKeepAliveSecond defines the connection time
	defaultKeepAliveSecond = 60 * time.Second
)

func init()  {
	client := &http.Client{}
	tp := &http.Transport{
		MaxIdleConns:        defaultMaxConnsPerHost,
		MaxIdleConnsPerHost: defaultMaxConnsPerHost,
		DialContext: (&net.Dialer{
			KeepAlive: defaultKeepAliveSecond,
			Timeout:   defaultTimeoutBySecond,
		}).DialContext,
	}
	tp.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true,
	}
	client.Transport = tp
	defalultClient = client
}

func NewDefaultHttpRequset(log ApolloLogInterface) *DefaultHttpRequest {
	defaultHttpRequest := new(DefaultHttpRequest)
	defaultHttpRequest.logger = log
	defaultHttpRequest.client = defalultClient
	return defaultHttpRequest
}


type HttpRequest interface {
	Request(url string) (*http.Response, error)
}

type DefaultHttpRequest struct {
	client *http.Client
	logger ApolloLogInterface
}

// 不带秘钥验证的请求方式
func (this *DefaultHttpRequest) Request(reqUrl string) (*http.Response, error) {
	client := this.client
	url, err := url2.Parse(reqUrl)
	if err != nil {
		this.logger.Error(fmt.Sprint("request Server url:%s, is invalid %s", url, err))
		return nil, err
	}
	var res *http.Response
	this.logger.Info(ParseUrl(reqUrl))
	req, err := http.NewRequest("GET", ParseUrl(reqUrl), nil)
	if req == nil || err != nil {
		return nil, errors.WithMessagef(err, "Generate connect request Fail,url:%s,Error:%s", reqUrl, err)
	}
	res, err = client.Do(req)
	if res == nil || err != nil {
		return nil, errors.WithMessagef(err, "Connect Server Fail,url:%s,Error:%s", reqUrl, err)
	}
	return res, err
}

// 对GET Url 做encode
func ParseUrl(uri string) string {
	url, _ := url2.Parse(uri)
	m,_ := url2.ParseQuery(url.RawQuery)
	buf := bytes.Buffer{}
	buf.WriteString(url.Scheme);
	buf.WriteString("://");
	buf.WriteString(url.Host);
	buf.WriteString(url.Path);
	querys := m.Encode()
	if querys != "" {
		buf.WriteString("?");
		buf.WriteString(querys);
	}
	return buf.String()
}
