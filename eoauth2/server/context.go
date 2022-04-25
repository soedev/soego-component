package server

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/soedev/soego-component/eoauth2/server/model"
	"github.com/soedev/soego/core/elog"
	"go.uber.org/zap"
)

// ResponseType enum
type ResponseType int

const (
	DATA ResponseType = iota
	REDIRECT
)

type Context struct {
	Ctx                context.Context
	responseType       ResponseType
	redirectUrl        string // 跳转地址
	responseErr        error  // 用户响应错误
	internalErr        error  // 用户内部错粗
	isError            bool
	redirectInFragment bool
	logger             *elog.Component
	output             ResponseData
	parentToken        model.Token // output会被设置到URL，自动生成的parent token，只能单独存储
}

// setRedirect changes the response to redirect to the given redirectUrl
func (c *Context) setRedirect(redirectUrl string) {
	// set redirect parameters
	c.responseType = REDIRECT
	c.redirectUrl = redirectUrl
}

// GetRedirectUrl returns the redirect redirectUrl with all query string parameters
func (c *Context) GetRedirectUrl() (string, error) {
	if c.responseType != REDIRECT {
		return "", errors.New("Not a redirect response")
	}

	u, err := url.Parse(c.redirectUrl)
	if err != nil {
		return "", err
	}

	var q url.Values
	if c.redirectInFragment {
		// start with empty set for fragment
		q = url.Values{}
	} else {
		// add parameters to existing query
		q = u.Query()
	}

	// add parameters
	for n, v := range c.output {
		q.Set(n, fmt.Sprint(v))
	}

	// https://tools.ietf.org/html/rfc6749#section-4.2.2
	// Fragment should be encoded as application/x-www-form-urlencoded (%-escaped, spaces are represented as '+')
	// The stdlib redirectUrl#String() doesn't make that easy to accomplish, so build this ourselves
	if c.redirectInFragment {
		u.Fragment = ""
		redirectURI := u.String() + "#" + q.Encode()
		return redirectURI, nil
	}

	// Otherwise, update the query and encode normally
	u.RawQuery = q.Encode()
	u.Fragment = ""
	return u.String(), nil
}

func (c *Context) setParentToken(token model.Token) {
	c.parentToken = token
}

func (c *Context) GetParentToken() model.Token {
	return c.parentToken
}

func (c *Context) SetOutput(key string, value interface{}) {
	c.output[key] = value
}

func (c *Context) GetOutput(key string) interface{} {
	return c.output[key]
}

func (c *Context) GetAllOutput() interface{} {
	return c.output
}

func (c *Context) IsError() bool {
	return c.isError
}

// setError sets an error id, description, and state on the Response
// uri is left blank
func (c *Context) setError(responseError string, internalError error, method string, description string) {
	// set error parameters
	c.isError = true
	c.responseErr = fmt.Errorf(responseError)
	if internalError != nil {
		// wrap error
		c.internalErr = fmt.Errorf(responseError+", %w", internalError)
	} else {
		c.internalErr = fmt.Errorf(responseError)
	}
	// 先取出state信息
	state := c.output["state"]
	c.output = make(ResponseData) // clear output
	c.output["error"] = c.responseErr.Error()
	c.output["state"] = state
	c.logger.Error("eoauth2_error", elog.FieldErr(c.internalErr), elog.FieldMethod(method), zap.String("description", description))
}

func (c *Context) setRedirectFragment(f bool) {
	c.redirectInFragment = f
}
