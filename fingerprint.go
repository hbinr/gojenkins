// Copyright 2015 Vadim Kravcenko
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package gojenkins

import (
	"context"
	"errors"
	"fmt"
)

type FingerPrint struct {
	Jenkins *Jenkins             `json:"jenkins"`
	Base    string               `json:"base"`
	Id      string               `json:"id"`
	Raw     *FingerPrintResponse `json:"raw"`
}

type FingerPrintResponse struct {
	FileName string `json:"fileName"`
	Hash     string `json:"hash"`
	Original struct {
		Name   string `json:"name"`
		Number int64  `json:"number"`
	} `json:"original"`
	Timestamp int64 `json:"timestamp"`
	Usage     []struct {
		Name   string `json:"name"`
		Ranges struct {
			Ranges []struct {
				End   int64 `json:"end"`
				Start int64 `json:"start"`
			} `json:"ranges"`
		} `json:"ranges"`
	} `json:"usage"`
}

func (f FingerPrint) Valid(ctx context.Context) (bool, error) {
	status, err := f.Poll(ctx)

	if err != nil {
		return false, err
	}

	if status != 200 || f.Raw.Hash != f.Id {
		return false, fmt.Errorf("Jenkins says %s is Invalid or the Status is unknown", f.Id)
	}
	return true, nil
}

func (f FingerPrint) ValidateForBuild(ctx context.Context, filename string, build *Build) (bool, error) {
	valid, err := f.Valid(ctx)
	if err != nil {
		return false, err
	}

	if valid {
		return true, nil
	}

	if f.Raw.FileName != filename {
		return false, errors.New("Filename does not Match")
	}
	if build != nil && f.Raw.Original.Name == build.Job.GetName() &&
		f.Raw.Original.Number == build.GetBuildNumber() {
		return true, nil
	}
	return false, nil
}

func (f FingerPrint) GetInfo(ctx context.Context) (*FingerPrintResponse, error) {
	_, err := f.Poll(ctx)
	if err != nil {
		return nil, err
	}
	return f.Raw, nil
}

func (f FingerPrint) Poll(ctx context.Context) (int, error) {
	response, err := f.Jenkins.Requester.GetJSON(ctx, f.Base+f.Id, f.Raw, nil)
	if err != nil {
		return 0, err
	}
	return response.StatusCode, nil
}
