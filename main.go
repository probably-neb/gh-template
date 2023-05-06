package main

import (
	"fmt"
	"github.com/cli/go-gh/v2/pkg/api"
    "strings"
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
    files := []GitFile{}
    for _, file := range response {
        if file.Type == "file" && strings.HasSuffix(file.Name, ".gitignore") {
            file.Name = strings.TrimSuffix(file.Name, ".gitignore")
            files = append(files, file)
        }
    }
    // TODO: get gitignores from community and global directories
    return files, nil
}

func main() {
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
		fmt.Printf("%s\n", file.Name)
	}
}
