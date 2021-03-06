//
// Copyright (c) 2018, Joyent, Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//

package compute_test

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/joyent/triton-go/compute"
	"github.com/joyent/triton-go/testutils"
)

var (
	mockVersions = []string{"7.0.0", "7.1.0", "7.2.0", "7.3.0", "8.0.0"}
	testError    = errors.New("unable to ping")
)

func TestPing(t *testing.T) {
	computeClient := MockComputeClient()

	do := func(ctx context.Context, pc *compute.ComputeClient) (*compute.PingOutput, error) {
		defer testutils.DeactivateClient()

		ping, err := pc.Ping(ctx)
		if err != nil {
			return nil, err
		}
		return ping, nil
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("GET", "/--ping", pingSuccessFunc)

		resp, err := do(context.Background(), computeClient)
		if err != nil {
			t.Fatal(err)
		}

		if resp.Ping != "pong" {
			t.Errorf("ping was not pong: expected %s", resp.Ping)
		}

		if !reflect.DeepEqual(resp.CloudAPI.Versions, mockVersions) {
			t.Errorf("ping did not contain CloudAPI versions: expected %s", mockVersions)
		}
	})

	t.Run("EOF decode", func(t *testing.T) {
		testutils.RegisterResponder("GET", "/--ping", pingEmptyFunc)

		_, err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "EOF") {
			t.Errorf("expected error to contain EOF: found %s", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("GET", "/--ping", pingErrorFunc)

		out, err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}
		if out != nil {
			t.Error("expected pingOut to be nil")
		}

		if !strings.Contains(err.Error(), "unable to ping") {
			t.Errorf("expected error to equal testError: found %s", err)
		}
	})

	t.Run("bad decode", func(t *testing.T) {
		testutils.RegisterResponder("GET", "/--ping", pingDecodeFunc)

		_, err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal(err)
		}

		if !strings.Contains(err.Error(), "invalid character") {
			t.Errorf("expected decode to fail: found %s", err)
		}
	})
}

func pingSuccessFunc(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{
	"ping": "pong",
	"cloudapi": {
		"versions": ["7.0.0", "7.1.0", "7.2.0", "7.3.0", "8.0.0"]
	}
}`)

	return &http.Response{
		StatusCode: 200,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func pingEmptyFunc(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	return &http.Response{
		StatusCode: 200,
		Header:     header,
		Body:       ioutil.NopCloser(strings.NewReader("")),
	}, nil
}

func pingErrorFunc(req *http.Request) (*http.Response, error) {
	return nil, testError
}

func pingDecodeFunc(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{
(ham!(//
}`)

	return &http.Response{
		StatusCode: 200,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}
