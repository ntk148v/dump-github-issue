// Copyright 2020 Kien Nguyen-Tuan <kiennt2609@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-github/github"
	"github.com/sethvargo/go-githubactions"
	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"golang.org/x/oauth2"
)

func main() {
	repository := githubactions.GetInput("repository")
	if repository == "" {
		githubactions.Fatalf("missing input 'repository'")
	}
	githubToken := githubactions.GetInput("github-token")
	if githubToken == "" {
		githubactions.Fatalf("missing input 'github-token'")
	}
	// Create github client
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)
	ownerRepo := strings.Split(repository, "/")
	// list all issues by repo
	issues, _, err := client.Issues.ListByRepo(ctx, ownerRepo[0], ownerRepo[1], &github.IssueListByRepoOptions{})
	if err != nil {
		githubactions.Fatalf("unable to list issues by repo: %s", err)
	}
	// markdown parser
	markdown := goldmark.New(
		goldmark.WithExtensions(extension.GFM, meta.Meta),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
		),
	)
	for _, issue := range issues {
		var buf bytes.Buffer
		context := parser.NewContext()
		if err := markdown.Convert([]byte(*issue.Body), &buf, parser.WithContext(context)); err != nil {
			githubactions.Fatalf("unable to convert issue body: %s", err)
		}
		metaData := meta.Get(context)
		if pathInt, ok := metaData["path"]; ok {
			path := pathInt.(string)
			if currContent, err := os.ReadFile(path); err != nil {
				if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
					githubactions.Fatalf("unable to create directory: %s", filepath.Dir(path))
				}
			} else {
				if strings.Compare(string(currContent), *issue.Body) == 0 {
					continue
				}
			}
			if err := os.WriteFile(path, []byte(*issue.Body), 0644); err != nil {
				githubactions.Fatalf("unable to write file: %s", path)
			}
		}
	}
}
