// Copyright 2016 Mender Software AS
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.
package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
  "time"
)

const correctUpdateResponse = `{\n
"image": {
"uri": "https://aws.my_update_bucket.com/kldjdaklj",
"checksum": "Hello, world!",
"id": "f81d4fae-7dec-11d0-a765-00a0c91e6bf6"
},
"id": "13876-123132-321123"
}`


func TestGetUpdate(t *testing.T) {

	// Test server that always responds with 200 code, and specific payload
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "application/json")
		//TODO
		fmt.Fprintln(w, correctUpdateResponse)
	}))
	defer ts.Close()

  client := setupTestClient(ts.URL)
  var config daemonConfigType
  config.setDeviceId()

  err, response := client.sendRequest(GET, ts.URL + "/" + config.deviceId + "/update")
  if err != nil {
    t.Fatal(err)
  }
  client.parseUpdateTesponse(response)
}


func TestCheckPeriodicDaemonUpdate(t *testing.T) {

  reqHandlingCnt := 0
  pullInterval := 1

	// Test server that always responds with 200 code, and specific payload
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "application/json")
		// we don't care about the payload here
		fmt.Fprintln(w, "OK")
    reqHandlingCnt += 1
	}))
	defer ts.Close()

  client := setupTestClient(ts.URL)
  var config daemonConfigType
  config.setPullInterval(pullInterval)
  config.setServerAddress(ts.URL)
  config.setDeviceId()

  go func() {
    runAsDemon(config, &client)
  }()

  timesPulled := 5
  time.Sleep(time.Duration(pullInterval * timesPulled) * time.Second)
  daemonQuit <- true

  if reqHandlingCnt < (timesPulled -1) {
    t.Fatal("Expected to receive at least ", timesPulled - 1, " requests - ", reqHandlingCnt, " received")
  }
}
