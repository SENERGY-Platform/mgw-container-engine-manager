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

package logger

import "encoding/json"

type Level uint8

func (l *Level) String() string {
	return lvlStrings[*l]
}

func (l *Level) UnmarshalJSON(data []byte) (err error) {
	var itf interface{}
	if err = json.Unmarshal(data, &itf); err != nil {
		return
	}
	*l, err = ParseLevel(itf.(string))
	return
}
