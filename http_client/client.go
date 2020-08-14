package steam_client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"

	httpSrv "github.com/CssHammer/go-template/http"
	"github.com/CssHammer/go-template/models"
)

const (
	RouteUser = "/user"
)

type Client struct {
	client    *http.Client
	userAgent string
	baseURL   string
}

func New(client *http.Client, userAgent string, baseURL string) *Client {
	return &Client{
		client:    client,
		userAgent: userAgent,
		baseURL:   baseURL,
	}
}

func (c *Client) GetUser(ctx context.Context, id int) (*models.User, error) {
	response := new(models.User)

	params := make(url.Values)
	params.Set(httpSrv.ParamID, fmt.Sprint(id))

	return response, c.request(ctx, fmt.Sprintf("%s/%d", RouteUser, id), params, nil, response)
}

func (c *Client) PostUser(ctx context.Context, user models.User) error {
	return c.request(ctx, RouteUser, nil, user, nil)
}

// request will make a call to the actual API and parse a response.
func (c *Client) request(ctx context.Context, requestURL string, params url.Values, body, response interface{}) error {
	// body
	var bodyBytes io.Reader
	var data []byte
	var err error
	if body != nil {
		data, err = json.Marshal(body)
		if err != nil {
			return err
		}
		bodyBytes = bytes.NewBuffer(data)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, bodyBytes)
	if err != nil {
		return err
	}

	// query params
	if params != nil {
		queryParams := req.URL.Query()
		for k, v := range params {
			for _, item := range v {
				queryParams.Add(k, item)
			}
		}
		req.URL.RawQuery = queryParams.Encode()
	}

	// headers
	req.Header.Set("User-Agent", c.userAgent)

	// response
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		// Do not unmarshall if response is nil
		if response == nil || reflect.ValueOf(response).IsNil() || len(data) == 0 {
			return nil
		}

		err = json.Unmarshal(data, response)
		if err != nil {
			return err
		}

		return nil
	}

	// This is an API Error
	return fmt.Errorf("api error: %v", string(data))
}
