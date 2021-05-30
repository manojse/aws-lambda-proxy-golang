package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func Handler(lambdactx context.Context, req map[string]interface{}) (events.APIGatewayProxyResponse, error) {
	// had issue with cors, added this for getting this fixed
	headers := map[string]string{
		"Access-Control-Allow-Origin": "*",
	}

	var request events.APIGatewayProxyRequest

	//marshal the request to json
	requestBytes, _ := json.Marshal(req)

	_ = json.Unmarshal(requestBytes, &request)
	fmt.Println("Request :", request)
	fmt.Println("APIRequestID : ", request.RequestContext.RequestID)
	queryString := url.Values{}
	for k, v := range request.QueryStringParameters {
		if k == "url" {
			continue
		}
		queryString.Add(k, v)
	}
	apiUrl := request.QueryStringParameters["url"]
	if len(queryString) > 0 {
		apiUrl = request.QueryStringParameters["url"] + "?" + queryString.Encode()
		if strings.Contains(request.QueryStringParameters["url"], "?") {
			apiUrl = request.QueryStringParameters["url"] + "&" + queryString.Encode()
		}
	}
	returnString, httpStatus, resHeader, _ := call(request.HTTPMethod, apiUrl, []byte(request.Body), request.Headers)

	for k, v := range resHeader {
		headers[k] = strings.Join(v, ",")
	}

	return events.APIGatewayProxyResponse{
		IsBase64Encoded: false,
		Body:            returnString,
		Headers:         headers,
		StatusCode:      httpStatus,
	}, nil
}

func main() {
	lambda.Start(Handler)
}

func call(method, url string, byteData []byte, header map[string]string) (pgResponse string, httpStatus int, resHeader map[string][]string, err error) {
	fmt.Println("URL", method, url)
	client := &http.Client{}
	req, err := http.NewRequest(method, url, bytes.NewBuffer(byteData))

	if err != nil {
		fmt.Println(err)
		return
	}

	delete(header, "Accept-Encoding")
	delete(header, "accept-encoding")
	delete(header, "Postman-Token")
	delete(header, "postman-token")
	delete(header, "Host")
	delete(header, "host")
	delete(header, "X-Forwarded-Proto")
	delete(header, "X-Forwarded-For")
	delete(header, "X-Forwarded-Port")
	delete(header, "X-Amzn-Trace-Id")
	delete(header, "User-Agent")
	delete(header, "x-forwarded-Pproto")
	delete(header, "x-forwarded-for")
	delete(header, "x-forwarded-port")
	delete(header, "x-amzn-Trace-id")
	delete(header, "user-agent")

	if _, ok := header["authorization"]; ok {
		header["Authorization"] = header["authorization"]
		delete(header, "authorization")
	}

	for k, v := range header {
		fmt.Println("header", k, " : ", v)
		req.Header.Add(k, v)
	}

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()
	httpStatus = res.StatusCode
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(res.Header)
	fmt.Println(string(body))
	return string(body), httpStatus, res.Header, nil
}
