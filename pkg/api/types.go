/*
Copyright 2023 The KubeSphere Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package api

type ListResult struct {
	Items      []interface{} `json:"items"`
	TotalItems int           `json:"totalItems"`
}

// NewListResult creates a ListResult for the given items and total.
func NewListResult(items []interface{}, total int) *ListResult {
	if items == nil {
		items = make([]interface{}, 0)
	}
	return &ListResult{
		Items:      items,
		TotalItems: total,
	}
}

const (
	StatusOK    = "ok"
	StatusError = "error"
)
