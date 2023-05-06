package main

import (
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
	Dir          string
}

var getAllGitignoreTemplates bool

func getEntriesFromDir(dir string, client *api.RESTClient, ch chan<- GitFile, wg *sync.WaitGroup) {
	defer wg.Done()
	response := []GitFile{}
	err := client.Get("repos/github/gitignore/contents/"+dir, &response)
	if err != nil {
		// return err
		return
	}

	for _, file := range response {
		if file.Type == "file" && strings.HasSuffix(file.Name, ".gitignore") {
			file.Name = strings.TrimSuffix(file.Name, ".gitignore")
			file.Dir = dir
			ch <- file
		} else if file.Type == "dir" && getAllGitignoreTemplates {
	wg.Add(1)
			getEntriesFromDir(dir+file.Name + "/", client, ch, wg)
		}
	}
}

func getGitignoreEntries(client *api.RESTClient) ([]GitFile, error) {
	ch := make(chan GitFile)
	var wg sync.WaitGroup;

	wg.Add(1)
	go getEntriesFromDir("", client, ch, &wg)

	go func() {
		wg.Wait()
		close(ch)
	}()

	files := []GitFile{}
	// block until ch is closed and all files have been recieved
	for file := range ch {
		files = append(files, file)
	}
	// TODO: get gitignores from community and global directories
	return files, nil
}

func main() {
	Cli := &cobra.Command{}

	IgnoreCmd := &cobra.Command{
		Use:     "gitignore",
		Example: "gh template gitignore --get python",
		Short:   "list and download the templates github provides for .gitignore files as well as licenses",
	}
	listFlag := IgnoreCmd.Flags().BoolP("list", "l", false, "list available .gitignore templates")
	getFlag := IgnoreCmd.Flags().StringP("get", "g", "", "get a gitignore template by name (use `--list` to list available templates)")
    IgnoreCmd.Flags().BoolVarP(&getAllGitignoreTemplates, "all", "b", false, "list community and global templates as well")
	IgnoreCmd.MarkFlagsMutuallyExclusive("get", "list")

	IgnoreCmd.Run = func(cmd *cobra.Command, args []string) {
		client, err := api.DefaultRESTClient()
		if err != nil {
			fmt.Println(err)
			return
		}

		if *listFlag {
			files, err := getGitignoreEntries(client)
			if err != nil {
				fmt.Println(err)
				return
			}
			for _, file := range files {
				fmt.Printf("%s%s\n", file.Dir, file.Name)
			}
		} else if cmd.Flags().Changed("get") {
            if !cmd.Flags().Changed("all") {
                getAllGitignoreTemplates = true
            }
			fmt.Println("get =", *getFlag)
		}
	}
	Cli.AddCommand(IgnoreCmd)
	err := Cli.Execute()
	if err != nil {
		fmt.Println(err)
	}
}
