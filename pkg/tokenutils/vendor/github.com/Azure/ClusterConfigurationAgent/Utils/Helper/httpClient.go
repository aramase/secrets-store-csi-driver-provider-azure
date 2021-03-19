package Helper

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
)

type HttpClient struct {
}

func (client *HttpClient) NewRequest(method, endpoint string, body io.Reader, headers map[string]string) (*http.Response, error) {

	tr := &http.Transport{
		Proxy:           http.ProxyFromEnvironment,
		TLSClientConfig: & tls.Config { MinVersion: tls.VersionTLS12},
	}

	log.Printf(fmt.Sprintf("URL:> %s Method:> %s", endpoint, method))

	req, err := http.NewRequest(method, endpoint, body)
	if err != nil {
		log.Printf("Error: in the HTTP. Method:{%v}, endpoint:{%v} error : {%v}", method, endpoint, err)
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")

	if headers != nil {
		//merging the headers
		for key, value := range headers {
			req.Header.Add(key, value)
		}
	}

	httpclient := &http.Client{Transport: tr}

	resp, err := httpclient.Do(req)
	return resp, err
}
