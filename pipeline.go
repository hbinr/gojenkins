// Copyright 2017 - Tessa Nordgren
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
//
// this file implements the pipeline-stage-view API:
// https://github.com/jenkinsci/pipeline-stage-view-plugin/tree/master/rest-api

package gojenkins

import (
	"context"
	"fmt"
	"regexp"
)

var baseURLRegex *regexp.Regexp

func init() {
	var err error
	baseURLRegex, err = regexp.Compile("(.+)/wfapi/.*$")
	if err != nil {
		panic(err)
	}
}

type PipelineRun struct {
	Job       *Job                         `json:"job"`
	Base      string                       `json:"base"`
	URLs      map[string]map[string]string `json:"_links"`
	ID        string                       `json:"id"`
	Name      string                       `json:"name"`
	Status    string                       `json:"status"`
	StartTime int64                        `json:"startTimeMillis"`
	EndTime   int64                        `json:"endTimeMillis"`
	Duration  int64                        `json:"durationMillis"`
	Stages    []PipelineNode               `json:"stages"`
}

type PipelineNode struct {
	Run            *PipelineRun                 `json:"run"`
	Base           string                       `json:"base"`
	URLs           map[string]map[string]string `json:"_links"`
	ID             string                       `json:"id"`
	Name           string                       `json:"name"`
	Status         string                       `json:"status"`
	StartTime      int64                        `json:"startTimeMillis"`
	Duration       int64                        `json:"durationMillis"`
	StageFlowNodes []PipelineNode               `json:"stageFlowNodes"`
	ParentNodes    []int64                      `json:"parentNodes"`
}

type PipelineInputAction struct {
	ID         string `json:"id"`
	Message    string `json:"message"`
	ProceedURL string `json:"proceedUrl"`
	AbortURL   string `json:"abortUrl"`
}

type PipelineArtifact struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Path string `json:"path"`
	URL  string `json:"url"`
	size int    `json:"size"`
}

type PipelineNodeLog struct {
	NodeID     string `json:"nodeId"`
	NodeStatus string `json:"nodeStatus"`
	Length     int64  `json:"length"`
	HasMore    bool   `json:"hasMore"`
	Text       string `json:"text"`
	ConsoleURL string `json:"consoleUrl"`
}

// utility function to fill in the Base fields under PipelineRun
func (run *PipelineRun) update() {
	href := run.URLs["self"]["href"]
	if matches := baseURLRegex.FindStringSubmatch(href); len(matches) > 1 {
		run.Base = matches[1]
	}
	for i := range run.Stages {
		run.Stages[i].Run = run
		href := run.Stages[i].URLs["self"]["href"]
		if matches := baseURLRegex.FindStringSubmatch(href); len(matches) > 1 {
			run.Stages[i].Base = matches[1]
		}
	}
}

func (job *Job) GetPipelineRuns(ctx context.Context) (pr []PipelineRun, err error) {
	_, err = job.Jenkins.Requester.GetJSON(ctx, job.Base+"/wfapi/runs", &pr, nil)
	if err != nil {
		return nil, err
	}
	for i := range pr {
		pr[i].update()
		pr[i].Job = job
	}

	return pr, nil
}

func (job *Job) GetPipelineRun(ctx context.Context, id string) (pr *PipelineRun, err error) {
	pr = new(PipelineRun)
	href := job.Base + "/" + id + "/wfapi/describe"
	_, err = job.Jenkins.Requester.GetJSON(ctx, href, pr, nil)
	if err != nil {
		return nil, err
	}
	pr.update()
	pr.Job = job

	return pr, nil
}

func (pr *PipelineRun) GetPendingInputActions(ctx context.Context) (PIAs []PipelineInputAction, err error) {
	PIAs = make([]PipelineInputAction, 0, 1)
	href := pr.Base + "/wfapi/pendingInputActions"
	_, err = pr.Job.Jenkins.Requester.GetJSON(ctx, href, &PIAs, nil)
	if err != nil {
		return nil, err
	}

	return PIAs, nil
}

func (pr *PipelineRun) GetArtifacts(ctx context.Context) (artifacts []PipelineArtifact, err error) {
	artifacts = make([]PipelineArtifact, 0, 0)
	href := pr.Base + "/wfapi/artifacts"
	_, err = pr.Job.Jenkins.Requester.GetJSON(ctx, href, artifacts, nil)
	if err != nil {
		return nil, err
	}

	return artifacts, nil
}

func (pr *PipelineRun) GetNode(ctx context.Context, id string) (node *PipelineNode, err error) {
	node = new(PipelineNode)
	href := pr.Base + "/execution/node/" + id + "/wfapi/describe"
	_, err = pr.Job.Jenkins.Requester.GetJSON(ctx, href, node, nil)
	if err != nil {
		return nil, err
	}

	return node, nil
}

func (node *PipelineNode) GetLog(ctx context.Context) (log *PipelineNodeLog, err error) {
	log = new(PipelineNodeLog)
	href := node.Base + "/wfapi/log"
	fmt.Println(href)
	_, err = node.Run.Job.Jenkins.Requester.GetJSON(ctx, href, log, nil)
	if err != nil {
		return nil, err
	}

	return log, nil
}
