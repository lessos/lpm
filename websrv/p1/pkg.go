// Copyright 2016 Eryx <evorui аt gmаil dοt cοm>, All rights reserved.
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

package p1 // import "github.com/sysinner/inpack/websrv/p1"

import (
	"net/http"
	"path/filepath"
	"sort"
	"strings"

	"github.com/hooto/httpsrv"
	"github.com/lessos/lessgo/types"

	"github.com/sysinner/inpack/ipapi"
	"github.com/sysinner/inpack/server/config"
	"github.com/sysinner/inpack/server/data"
)

type Pkg struct {
	*httpsrv.Controller
}

func (c Pkg) DlAction() {

	c.AutoRender = false

	file := filepath.Clean(c.Request.RequestPath)

	if !strings.HasPrefix(file, "ips/p1/pkg/dl/") {
		c.RenderError(400, "Bad Request")
		return
	}

	// TODO auth
	fs_dir := config.Config.StorageConnect.Value("data_dir")
	if fs_dir == "" {
		c.RenderError(400, "Bad Request")
		return
	}

	http.ServeFile(
		c.Response.Out,
		c.Request.Request,
		fs_dir+file[len("ips/p1/pkg/dl"):],
	)
}

func (c Pkg) ListAction() {

	ls := ipapi.PackageList{}
	defer c.RenderJson(&ls)

	var (
		q_name    = c.Params.Get("name")
		q_channel = c.Params.Get("channel")
		q_text    = c.Params.Get("q")
		limit     = int(c.Params.Int64("limit"))
	)

	if !ipapi.PackageNameRe.MatchString(q_name) {
		ls.Error = types.NewErrorMeta("400", "Invalid Package Name")
		return
	}

	if limit < 1 {
		limit = 100
	} else if limit > 200 {
		limit = 200
	}

	var (
		offset = ipapi.DataPackKey(q_name)
		cutset = ipapi.DataPackKey(q_name)
	)

	rs := data.Data.NewReader(nil).KeyRangeSet(offset, cutset).LimitNumSet(1000).Query()
	if !rs.OK() {
		ls.Error = types.NewErrorMeta("500", rs.Message)
		return
	}

	for _, entry := range rs.Items {

		if len(ls.Items) >= limit {
			// TOPO return 0
		}

		var set ipapi.Package
		if err := entry.Decode(&set); err != nil {
			continue
		}

		if q_name != "" && q_name != set.Meta.Name {
			continue
		}

		if q_channel != "" && q_channel != set.Channel {
			continue
		}

		if q_text != "" && !strings.Contains(set.Meta.Name, q_text) {
			continue
		}

		if ipapi.OpPermAllow(set.OpPerm, ipapi.OpPermOff) {
			continue
		}

		ls.Items = append(ls.Items, set)
	}

	sort.Slice(ls.Items, func(i, j int) bool {
		return ls.Items[i].Meta.Updated > ls.Items[j].Meta.Updated
	})

	if len(ls.Items) > limit {
		ls.Items = ls.Items[:limit]
	}

	ls.Kind = "PackageList"
}

func (c Pkg) EntryAction() {

	var set struct {
		types.TypeMeta
		ipapi.Package
	}
	defer c.RenderJson(&set)

	var (
		id   = c.Params.Get("id")
		name = c.Params.Get("name")
	)

	if id == "" && name == "" {
		set.Error = types.NewErrorMeta("400", "Package ID or Name Not Found")
		return
	} else if name != "" {

		if !ipapi.PackageNameRe.MatchString(name) {
			set.Error = types.NewErrorMeta("400", "Invalid Package Name")
			return
		}

		version := ipapi.PackageVersion{
			Version: types.Version(c.Params.Get("version")),
			Release: types.Version(c.Params.Get("release")),
			Dist:    c.Params.Get("dist"),
			Arch:    c.Params.Get("arch"),
		}
		if err := version.Valid(); err != nil {
			set.Error = types.NewErrorMeta("400", err.Error())
			return
		}
		id = ipapi.PackageFilenameKey(name, version)
	}

	if id != "" {

		if rs := data.Data.NewReader(ipapi.DataPackKey(id)).Query(); rs.OK() {
			rs.Decode(&set.Package)
		} else if name != "" {
			// TODO
		} else {

		}
	}

	if set.Meta.Name == "" {
		set.Error = types.NewErrorMeta("404", "Package Not Found")
	} else {
		set.Kind = "Package"
	}
}
