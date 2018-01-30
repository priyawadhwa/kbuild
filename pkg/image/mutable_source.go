/*
Copyright 2017 Google, Inc. All rights reserved.
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

package image

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"time"

	"github.com/containers/image/manifest"
	"github.com/containers/image/types"
	digest "github.com/opencontainers/go-digest"
)

type MutableSource struct {
	ProxySource
	mfst        *manifest.Schema2
	cfg         *manifest.Schema2Image
	extraBlobs  map[string][]byte
	extraLayers []digest.Digest
}

func NewMutableSource(r types.ImageReference) (*MutableSource, error) {
	src, err := r.NewImageSource(nil)
	if err != nil {
		return nil, err
	}
	img, err := r.NewImage(nil)
	if err != nil {
		return nil, err
	}

	ms := &MutableSource{
		ProxySource: ProxySource{
			Ref: r,
			src: src,
			img: img,
		},
		extraBlobs: make(map[string][]byte),
	}
	if err := ms.populateManifestAndConfig(); err != nil {
		return nil, err
	}
	return ms, nil
}

// GetManifest marshals the stored manifest to the byte format.
func (m *MutableSource) GetManifest(instanceDigest *digest.Digest) ([]byte, string, error) {
	s, err := json.Marshal(m.mfst)
	if err := m.saveConfig(); err != nil {
		return nil, "", err
	}
	return s, manifest.DockerV2Schema2MediaType, err
}

// populateManifestAndConfig parses the raw manifest and configs, storing them on the struct.
func (m *MutableSource) populateManifestAndConfig() error {
	mfstBytes, _, err := m.src.GetManifest(nil)
	if err != nil {
		return err
	}

	m.mfst, err = manifest.Schema2FromManifest(mfstBytes)
	fmt.Println("Manifest is ", m.mfst)
	if err != nil {
		return err
	}

	bi := types.BlobInfo{Digest: m.mfst.ConfigDescriptor.Digest}
	r, _, err := m.src.GetBlob(bi)
	if err != nil {
		return err
	}

	cfgBytes, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	return json.Unmarshal(cfgBytes, &m.cfg)
}

// GetBlob first checks the stored "extra" blobs, then proxies the call to the original source.
func (m *MutableSource) GetBlob(bi types.BlobInfo) (io.ReadCloser, int64, error) {
	if b, ok := m.extraBlobs[bi.Digest.String()]; ok {
		return ioutil.NopCloser(bytes.NewReader(b)), int64(len(b)), nil
	}
	return m.src.GetBlob(bi)
}

func gzipBytes(b []byte) ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	w := gzip.NewWriter(buf)
	_, err := w.Write(b)
	w.Close()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// AppendLayer appends an uncompressed blob to the image, preserving the invariants required across the config and manifest.
func (m *MutableSource) AppendLayer(content []byte) error {
	diffID := digest.FromBytes(content)

	// Add the layer to the manifest.
	descriptor := manifest.Schema2Descriptor{
		MediaType: manifest.DockerV2Schema2LayerMediaType,
		Size:      int64(len(content)),
		Digest:    diffID,
	}
	m.mfst.LayersDescriptors = append(m.mfst.LayersDescriptors, descriptor)

	m.extraBlobs[diffID.String()] = content
	m.extraLayers = append(m.extraLayers, diffID)

	// Also add it to the config.
	m.cfg.RootFS.DiffIDs = append(m.cfg.RootFS.DiffIDs, diffID)
	history := manifest.Schema2History{
		Created: time.Now(),
		Author:  "kbuild",
	}
	m.cfg.History = append(m.cfg.History, history)

	return nil
}

// WriteManifest writes the final manfiest to a file at path
func (m *MutableSource) WriteManifest(path string) error {
	mfstContents, err := m.mfst.Serialize()
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(path, mfstContents, 0644)
	return err
}

func (m *MutableSource) WriteConfig(path string) error {
	cfgBlob, err := json.Marshal(m.cfg)
	if err != nil {
		return err
	}
	cfgDigest := digest.FromBytes(cfgBlob).String()
	d := strings.Split(cfgDigest, ":")[1]
	filePath := path + d + ".tar"
	err = ioutil.WriteFile(filePath, cfgBlob, 0644)
	m.saveConfig()
	return err
}

// saveConfig marshals the stored image config, and updates the references to it in the manifest.
func (m *MutableSource) saveConfig() error {
	cfgBlob, err := json.Marshal(m.cfg)
	if err != nil {
		return err
	}

	cfgDigest := digest.FromBytes(cfgBlob)
	m.extraBlobs[cfgDigest.String()] = cfgBlob
	m.mfst.ConfigDescriptor = manifest.Schema2Descriptor{
		MediaType: manifest.DockerV2Schema2ConfigMediaType,
		Size:      int64(len(cfgBlob)),
		Digest:    cfgDigest,
	}
	return nil
}
