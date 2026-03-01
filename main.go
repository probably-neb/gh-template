package main

import (
	"encoding/base64"
	"fmt"
	"strings"
	"sync"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/spf13/cobra"
)

type GitFile struct {
	Type         string
	Name         string
	Download_Url string
	Path         string
	Content      string
	Encoding     string
}

func getEntriesFromDir(dir string, client *api.RESTClient, ch chan<- GitFile, wg *sync.WaitGroup) {
	defer wg.Done()
	response := []GitFile{}
	err := client.Get("repos/github/gitignore/contents/"+dir, &response)
	if err != nil {
		// TODO: return err
		return
	}

	for _, file := range response {
		ch <- file
	}
}

func ListGitignoreTemplates(client *api.RESTClient, recurse bool) ([]GitFile, error) {
	ch := make(chan GitFile)
	var wg sync.WaitGroup

	wg.Add(1)
	go getEntriesFromDir("", client, ch, &wg)

	go func() {
		wg.Wait()
		close(ch)
	}()

	files := []GitFile{}
	// block until ch is closed and all files have been recieved
	for file := range ch {
		if file.Type == "file" && strings.HasSuffix(file.Name, ".gitignore") {
			files = append(files, file)
		} else if file.Type == "dir" && recurse {
			wg.Add(1)
			go getEntriesFromDir(file.Path, client, ch, &wg)
		}
	}
	return files, nil
}

func main() {
	Cli := &cobra.Command{}

    // TODO: License command very similar to Ignore command
    // TODO: complete subcommand for generating 
	IgnoreCmd := &cobra.Command{
		Use:     "gitignore",
        Aliases: []string{"ignore"},
		Example: "gh template gitignore --get Python",
		Short:   "list and download the templates github provides for .gitignore files as well as licenses",
	}
    // TODO: merge command for merging list of gitignore files into one (use special argument to merge with current repos .gitignore as well)
	listFlag := IgnoreCmd.Flags().BoolP("list", "l", false, "list available .gitignore templates")
	getFlag := IgnoreCmd.Flags().StringP("get", "g", "", "get a gitignore template by name (use `--list` to list available templates)")
	listAllFlag := IgnoreCmd.Flags().BoolP("all", "b", false, "list community and global templates as well. For explanations of what these two groups are see https://github.com/github/gitignore#folder-structure")
	IgnoreCmd.MarkFlagsMutuallyExclusive("get", "list")

	IgnoreCmd.Run = func(cmd *cobra.Command, args []string) {
		client, err := api.DefaultRESTClient()
		if err != nil {
			fmt.Println(err)
			return
		}

		if *listFlag {
			templates, err := ListGitignoreTemplates(client, *listAllFlag)
			if err != nil {
				fmt.Println(err)
				return
			}
			for _, template := range templates {
				fmt.Printf("%s\n", strings.TrimSuffix(template.Path, ".gitignore"))
			}
		} else if cmd.Flags().Changed("get") {
            name := *getFlag
			if !strings.HasSuffix(name, ".gitignore") {
				name += ".gitignore"
			}
			var template GitFile
            url := "repos/github/gitignore/contents/"+name
			err := client.Get(url, &template)
			if err != nil {
				fmt.Println(err)
				return
			}
			if template.Encoding == "base64" {
				// TODO: is contents always included?
				contents, err := base64.StdEncoding.DecodeString(template.Content)
				if err != nil {
					fmt.Println(err)
					return
				}
				fmt.Printf("%s\n", contents)
			} else {
				fmt.Println("unknown encoding:", template.Encoding)
				return
			}
		}
	}
	Cli.AddCommand(IgnoreCmd)
	err := Cli.Execute()
	if err != nil {
		fmt.Println(err)
	}
}
