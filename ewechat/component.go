// Copyright 2020
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ewechat

import (
	"github.com/soedev/soego/core/elog"

	"github.com/soedev/soego-component/ewechat/cache"
	"github.com/soedev/soego-component/ewechat/context"
	"github.com/soedev/soego-component/ewechat/miniprogram"
)

type Component struct {
	config *config
	ctx    *context.Context
	client cache.Cache
	logger *elog.Component
}

func newComponent(cfg *config, ctx *context.Context, client cache.Cache, logger *elog.Component) *Component {
	return &Component{
		config: cfg,
		ctx:    ctx,
		client: client,
		logger: logger,
	}
}

// GetMiniProgram 获取小程序的实例
func (c *Component) GetMiniProgram() *miniprogram.MiniProgram {
	return miniprogram.NewMiniProgram(c.ctx)
}
