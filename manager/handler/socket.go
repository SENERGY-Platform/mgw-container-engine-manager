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

package handler

import (
	"deployment-manager/util/logger"
	"net"
	"os"
)

func clean(path string) (err error) {
	if _, err = os.Stat(path); err == nil {
		return os.Remove(path)
	} else if os.IsNotExist(err) {
		err = nil
	}
	return
}

func NewUnixListener(path string) (listener *net.UnixListener, err error) {
	if err = clean(path); err != nil {
		return
	}
	if listener, err = net.ListenUnix("unix", &net.UnixAddr{Name: path, Net: "unix"}); err != nil {
		return
	}
	defer func() {
		if err != nil {
			if err = listener.Close(); err != nil {
				logger.Error(err)
			}
		}
	}()
	if os.Getuid() == 0 {
		if err = os.Chown(path, os.Getuid(), 101); err != nil {
			return
		}
		if err = os.Chmod(path, 0660); err != nil {
			return
		}
	} else {
		if err = os.Chmod(path, 0666); err != nil {
			return
		}
	}
	return
}
