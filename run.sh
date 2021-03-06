#!/bin/bash
#
# Copyright 2020 Aletheia Ware LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -e
set -x

go fmt $GOPATH/src/github.com/AletheiaWareLLC/conveyservergo
go vet $GOPATH/src/github.com/AletheiaWareLLC/conveyservergo
go test $GOPATH/src/github.com/AletheiaWareLLC/conveyservergo
export BETA=true
export LIVE=false
export HTTPS=false
export ROOT_DIRECTORY=.
export ALIAS=test-convey
export PASSWORD=password1234
source private/config.sh
go run github.com/AletheiaWareLLC/conveyservergo start $@
