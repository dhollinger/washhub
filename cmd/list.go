// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/google/go-github/github"
	"github.com/spf13/cobra"
	"github.com/tidwall/pretty"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list <path> [<state>]",
	Short: "List content at <path>",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		path := strings.TrimSuffix(args[0], "/")
		if path == "/github" { // we are at top level
			listOrganisations()
		} else {
			path = strings.TrimPrefix(path, "/github/")
			if strings.Contains(path, "/") { // we have at least org and repo in the path
				listContent(path)
			} else { // we have only an org in the path
				listRepos(path)
			}
			os.Exit(0)
		}
		os.Exit(0)
	},
}

type attributes struct {
	Size int `json:"size"`
}

type entry struct {
	Name       string     `json:"name"`
	Methods    []string   `json:"methods"`
	Attributes attributes `json:"attributes"`
}

func listContent(path string) {
	org, tail := SplitPath(path)
	repo, filePath := SplitPath(tail)
	_, dirContent, err := FetchRepositoryContent(org, repo, filePath)
	HandleError(err)
	entries := directoryToEntries(dirContent)
	PrintEntries(entries)
}

func listRepos(org string) {
	repos, _, err := GithubClient().Repositories.List(context.Background(), org, nil)
	HandleError(err)
	entries := reposToEntries(repos)
	PrintEntries(entries)
}

func listOrganisations() {
	orgs, _, err := GithubClient().Organizations.ListOrgMemberships(context.Background(), nil)
	HandleError(err)

	var entries []*entry

	userEntry := &entry{
		Name:    UserName(),
		Methods: []string{"list"},
	}

	entries = append(entries, userEntry)

	for _, org := range orgs {
		orgEntry := &entry{
			Name:    *org.GetOrganization().Login,
			Methods: []string{"list"},
		}
		entries = append(entries, orgEntry)
	}
	PrintEntries(entries)
	os.Exit(0)
}

// PrintEntries prints the entries
func PrintEntries(entries []*entry) {
	json, _ := json.Marshal(entries)
	fmt.Println(string(pretty.Pretty(json)))
}

// PrintEntry prints one entry
func PrintEntry(entry *entry) {
	json, _ := json.Marshal(entry)
	fmt.Println(string(pretty.Pretty(json)))
}

// directoryToEntries converts github dirContent to entries
func directoryToEntries(dirEntries []*github.RepositoryContent) []*entry {
	entries := make([]*entry, len(dirEntries))
	for i, dirEntry := range dirEntries {
		myEntry := &entry{
			Name: *dirEntry.Name,
			Attributes: attributes{
				Size: *dirEntry.Size,
			},
		}
		if *dirEntry.Type == "file" {
			myEntry.Methods = []string{"read"}
		} else {
			myEntry.Methods = []string{"list"}
		}

		entries[i] = myEntry
	}
	return entries

}

// reposToEntries converts repos to entries
func reposToEntries(repos []*github.Repository) []*entry {
	entries := make([]*entry, len(repos))
	for i, repo := range repos {
		myEntry := &entry{
			Name:    *repo.Name,
			Methods: []string{"list"},
		}
		entries[i] = myEntry
	}
	return entries
}

func init() {
	rootCmd.AddCommand(listCmd)
}
