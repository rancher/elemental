/*
Copyright © 2022 SUSE LLC

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

package tpm

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/google/certificate-transparency-go/x509"
	"github.com/google/go-attestation/attest"

	gotpm "github.com/rancher-sandbox/go-tpm"

	"github.com/gorilla/websocket"
	v1 "github.com/rancher-sandbox/os2/pkg/apis/rancheros.cattle.io/v1"
	"github.com/rancher/wrangler/pkg/merr"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (a *AuthServer) verifyChain(ek *attest.EK, namespace string) error {
	secret, err := a.secretCache.Get(namespace, tpmCACert)
	if apierrors.IsNotFound(err) {
		return nil
	}

	roots := x509.NewCertPool()
	_ = roots.AppendCertsFromPEM(secret.Data[corev1.TLSCertKey])
	opts := x509.VerifyOptions{
		Roots: roots,
	}
	_, err = ek.Certificate.Verify(opts)
	return err
}

func (a *AuthServer) validHash(ek *attest.EK, registerNamespace string) (*v1.MachineInventory, error) {
	hashEncoded, err := gotpm.DecodePubHash(ek)
	if err != nil {
		return nil, fmt.Errorf("tpm: could not get public key hash: %v", err)
	}

	if registerNamespace != "" {
		if err := a.verifyChain(ek, registerNamespace); err != nil {
			return nil, fmt.Errorf("verifying chain: %w", err)
		}
		return &v1.MachineInventory{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: registerNamespace,
			},
			Spec: v1.MachineInventorySpec{
				TPMHash: hashEncoded,
			},
		}, nil
	}

	machines, err := a.machineCache.GetByIndex(machineByHash, hashEncoded)
	if apierrors.IsNotFound(err) || len(machines) != 1 {
		if len(machines) > 1 {
			logrus.Errorf("multiple machines for same hash %s found: %v", hashEncoded, machines)
		}
		return nil, fmt.Errorf("failed to find machine")
	}

	if err := a.verifyChain(ek, machines[0].Namespace); err != nil {
		return nil, fmt.Errorf("verifying chain: %w", err)
	}

	return machines[0], nil
}

func writeRead(conn *websocket.Conn, input []byte) ([]byte, error) {
	writer, err := conn.NextWriter(websocket.BinaryMessage)
	if err != nil {
		return nil, err
	}

	if _, err := writer.Write(input); err != nil {
		return nil, err
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	_, reader, err := conn.NextReader()
	if err != nil {
		return nil, err
	}

	return ioutil.ReadAll(reader)
}

func upgrade(resp http.ResponseWriter, req *http.Request) (*websocket.Conn, error) {
	upgrader := websocket.Upgrader{
		HandshakeTimeout: 5 * time.Second,
		CheckOrigin:      func(r *http.Request) bool { return true },
	}

	conn, err := upgrader.Upgrade(resp, req, nil)
	if err != nil {
		return nil, err
	}
	_ = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	_ = conn.SetReadDeadline(time.Now().Add(10 * time.Second))

	return conn, err
}

func (a *AuthServer) Authenticate(resp http.ResponseWriter, req *http.Request, registerNamespace string) (*v1.MachineInventory, bool, io.WriteCloser, error) {
	header := req.Header.Get("Authorization")
	if !strings.HasPrefix(header, "Bearer TPM") {
		return nil, true, nil, nil
	}

	ek, attestationData, err := gotpm.GetAttestationData(header)
	if err != nil {
		return nil, false, nil, err
	}

	machine, err := a.validHash(ek, registerNamespace)
	if err != nil {
		return nil, false, nil, err
	}

	secret, challenge, err := gotpm.GenerateChallenge(ek, attestationData)
	if err != nil {
		return nil, false, nil, err
	}

	conn, err := upgrade(resp, req)
	if err != nil {
		return nil, false, nil, err
	}

	challResp, err := writeRead(conn, challenge)
	if err != nil {
		return nil, false, nil, err
	}

	if err := gotpm.ValidateChallenge(secret, challResp); err != nil {
		return nil, false, nil, err
	}

	writer, err := conn.NextWriter(websocket.BinaryMessage)
	return machine, false, &responseWriter{
		WriteCloser: writer,
		conn:        conn,
	}, err
}

type responseWriter struct {
	io.WriteCloser
	conn *websocket.Conn
}

func (r *responseWriter) Close() error {
	err := r.WriteCloser.Close()
	err2 := r.conn.Close()
	return merr.NewErrors(err, err2)
}
