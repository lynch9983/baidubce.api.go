package api

import (
	"encoding/base64"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	jsoniter "github.com/json-iterator/go"
)

type Api interface {
	// 识别图片上的文字(通用)
	ImageToString(img []byte) (text string, err error)
	// 识别图片上的数字
	ImageToNum(img []byte) (text string, err error)
}

type api struct {
	apiKey    string
	secretKey string

	refresh_token string
	expires_in    time.Time
	access_token  string
}

func NewApi(apiKey string, secretKey string) Api {
	c := &api{}
	c.apiKey = apiKey
	c.secretKey = secretKey

	return c
}

type wordsRec struct {
	Words string `json:"words"`
}

type wordsData struct {
	Count   int        `json:"words_result_num"`
	Records []wordsRec `json:"words_result"`
}

func (c *api) ImageToString(img []byte) (text string, err error) {
	err = c.checkToken()
	if err != nil {
		return text, err
	}

	uri := "https://aip.baidubce.com/rest/2.0/ocr/v1/general?access_token=" + c.access_token
	params := url.Values{}
	params.Add("image", base64.StdEncoding.EncodeToString(img))
	resp, err := http.PostForm(uri, params)
	if err != nil {
		return text, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return text, err
	}
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	var res wordsData
	json.Unmarshal(body, &res)
	if res.Count == 0 || res.Records[0].Words == "" {
		return text, errors.New("未识别出文字")
	}

	text = res.Records[0].Words

	return text, err
}

func (c *api) ImageToNum(img []byte) (text string, err error) {
	err = c.checkToken()
	if err != nil {
		return text, err
	}

	uri := "https://aip.baidubce.com/rest/2.0/ocr/v1/numbers?access_token=" + c.access_token
	params := url.Values{}
	params.Add("image", base64.StdEncoding.EncodeToString(img))
	resp, err := http.PostForm(uri, params)
	if err != nil {
		return text, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return text, err
	}
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	var res wordsData
	json.Unmarshal(body, &res)
	if res.Count == 0 || res.Records[0].Words == "" {
		return text, errors.New("未识别出文字")
	}

	text = res.Records[0].Words

	return text, err
}

type errData struct {
	Error_name        string `json:"error"`
	Error_description string `json:"error_description"`
}

type tokenData struct {
	errData
	Refresh_token string `json:"refresh_token"`
	Expires_in    int64  `json:"expires_in"`
	Access_token  string `json:"access_token"`
}

func (c *api) checkToken() error {

	if c.access_token != "" && time.Now().Before(c.expires_in) {
		return nil
	}

	uri := "https://aip.baidubce.com/oauth/2.0/token"
	params := url.Values{}
	params.Add("grant_type", "client_credentials")
	params.Add("client_id", c.apiKey)
	params.Add("client_secret", c.secretKey)
	resp, err := http.PostForm(uri, params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	var res tokenData
	json.Unmarshal(body, &res)
	if res.Error_name != "" {
		return errors.New(res.Error_description)
	}

	c.refresh_token = res.Refresh_token
	c.expires_in = time.Now().Add(time.Duration(res.Expires_in) * time.Second)
	c.access_token = res.Access_token

	return nil
}
