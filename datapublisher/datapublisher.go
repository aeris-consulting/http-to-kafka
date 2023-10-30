//    Copyright 2021 AERIS-Consulting e.U.
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package datapublisher

type DataPublisher interface {
	Publish(key []byte, message []byte)
	PublishToDestination(destination string, key []byte, message []byte)
}

var (
	RegisteredPublishers []DataPublisher
)

// RegisterPublisher registers a new DataPublisher to use when a request with data is received.
func RegisterPublisher(publisher DataPublisher) {
	RegisteredPublishers = append(RegisteredPublishers, publisher)
}
