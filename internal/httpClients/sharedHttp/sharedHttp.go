// Copyright 2024 lukemakhanu
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

package sharedHttp

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type SharedHttpConfRepository interface {
	JSON(w http.ResponseWriter, code int, src interface{})
	Error(w http.ResponseWriter, code int, err error, msg string)
	CORSMiddleware() gin.HandlerFunc
	SecurityMiddleware(expectedHost string) gin.HandlerFunc
}
