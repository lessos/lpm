// Copyright 2016 lessos Authors, All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1 // import "code.hooto.com/lessos/lospack/websrv/v1"

import (
	"sort"
	"strings"

	"code.hooto.com/lessos/lospack/lpapi"
	"code.hooto.com/lessos/lospack/server/data"
	"code.hooto.com/lynkdb/iomix/skv"
	"github.com/lessos/lessgo/httpsrv"
	"github.com/lessos/lessgo/types"
)

type PkgInfo struct {
	*httpsrv.Controller
}

func (c PkgInfo) ListAction() {

	sets := lpapi.PackageInfoList{}
	defer c.RenderJson(&sets)

	var (
		qry_text = c.Params.Get("qry_text")
		limit    = 100
	)

	rs := data.Data.PvScan("info/", "", "", 10000)
	if !rs.OK() {
		sets.Error = &types.ErrorMeta{
			Code:    "500",
			Message: "Server Error",
		}
		return
	}

	rs.KvEach(func(entry *skv.ResultEntry) int {

		if len(sets.Items) > limit {
			return -1
		}

		var set lpapi.PackageInfo
		if err := entry.Decode(&set); err == nil {

			if qry_text != "" && !strings.Contains(set.Meta.Name, qry_text) {
				return 0
			}

			sets.Items = append(sets.Items, set)
		}

		return 0
	})

	sort.Slice(sets.Items, func(i, j int) bool {
		return sets.Items[i].Meta.Updated > sets.Items[j].Meta.Updated
	})

	sets.Kind = "PackageInfoList"
}

func (c PkgInfo) EntryAction() {

	set := lpapi.PackageInfo{}
	defer c.RenderJson(&set)

	if c.Params.Get("name") == "" {
		set.Error = &types.ErrorMeta{
			Code:    "404",
			Message: "PackageInfo Not Found",
		}
		return
	}

	rs := data.Data.PvGet("info/" + c.Params.Get("name"))
	if !rs.OK() {
		set.Error = &types.ErrorMeta{
			Code:    "404",
			Message: "PackageInfo Not Found",
		}
		return
	}

	if err := rs.Decode(&set); err != nil {
		set.Error = &types.ErrorMeta{
			Code:    "404",
			Message: "PackageInfo Not Found",
		}
		return
	}

	set.Kind = "PackageInfo"
}

func (c PkgInfo) SetAction() {

	set := lpapi.PackageInfo{}
	defer c.RenderJson(&set)

	if err := c.Request.JsonDecode(&set); err != nil {
		set.Error = &types.ErrorMeta{
			Code:    "400",
			Message: "Bad Request",
		}
		return
	}

	if rs := data.Data.PvGet("info/" + set.Meta.Name); !rs.OK() {
		set.Error = &types.ErrorMeta{
			Code:    "400",
			Message: "PackageInfo Not Found",
		}
		return
	} else {

		var prev lpapi.PackageInfo

		if err := rs.Decode(&prev); err != nil {
			set.Error = &types.ErrorMeta{
				Code:    "500",
				Message: "Server Error",
			}
			return
		}

		if prev.Description != set.Description {
			prev.Description = set.Description
		}

		prev.Kind = ""
		if rs := data.Data.PvPut("info/"+set.Meta.Name, prev, nil); !rs.OK() {
			set.Error = &types.ErrorMeta{
				Code:    "500",
				Message: "Server Error",
			}
			return
		}
	}

	set.Kind = "PackageInfo"
}