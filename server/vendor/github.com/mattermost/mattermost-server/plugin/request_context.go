// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package plugin

type Context struct {
}

func NewBlankContext() *Context {
	return &Context{}
}
