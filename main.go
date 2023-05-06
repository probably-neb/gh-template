package main

import (
	"fmt"
	"github.com/cli/go-gh/v2/pkg/api"
)

type GitFile struct {
	Type         string
	Name         string
	Download_Url string
}

func getGitignoreEntries(client *api.RESTClient) ([]GitFile,error) {
    response := []GitFile{}
    err := client.Get("repos/github/gitignore/contents/", &response)
	if err != nil {
		return nil, err
	}
    // TODO: get gitignores from community and global directories
    return response, nil
}

func main() {
	fmt.Println("hi world, this is the gh-template extension!")
	client, err := api.DefaultRESTClient()
	if err != nil {
		fmt.Println(err)
		return
	}
    files, err := getGitignoreEntries(client)
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, file := range files {
		fmt.Printf("%+v\n", file)
	}
}

// For more examples of using go-gh, see:
// https://github.com/cli/go-gh/blob/trunk/example_gh_test.go
