package main

import (
	"fmt"
	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/spf13/cobra"
	"strings"
)

type GitFile struct {
	Type         string
	Name         string
	Download_Url string
}

func getGitignoreEntries(client *api.RESTClient) ([]GitFile, error) {
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
	Cli := &cobra.Command{}

	IgnoreCmd := &cobra.Command{
		Use:     "gitignore",
		Example: "gh template gitignore --get python",
		Short:   "list and download the templates github provides for .gitignore files as well as licenses",
	}
	listFlag := IgnoreCmd.Flags().BoolP("list", "l", false, "list available .gitignore templates")
	getFlag := IgnoreCmd.Flags().StringP("get", "g", "", "get a gitignore template by name (use `--list` to list available templates)")
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
				fmt.Printf("%s\n", file.Name)
			}
		} else if cmd.Flags().Changed("get") {
			fmt.Println("get =", *getFlag)
		}
	}
	Cli.AddCommand(IgnoreCmd)
	err := Cli.Execute()
	if err != nil {
		fmt.Println(err)
	}
}
