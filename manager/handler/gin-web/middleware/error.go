/*
 * Copyright 2022 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package middleware

import (
	"deployment-manager/manager/api"
	"deployment-manager/manager/itf"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
)

func ErrorHandler(gc *gin.Context) {
	gc.Next()
	if !gc.IsAborted() && len(gc.Errors) > 0 {
		if gc.Writer.Status() < 400 {
			gc.Status(http.StatusInternalServerError)
		}
		var errs []string
		for _, e := range gc.Errors {
			var ceErr *itf.Error
			if errors.As(e, &ceErr) {
				gc.Status(ceErr.Code())
			}
			errs = append(errs, e.Error())
		}
		gc.JSON(-1, &api.ErrResponse{
			Status: http.StatusText(gc.Writer.Status()),
			Errors: errs,
		})
	}
}
