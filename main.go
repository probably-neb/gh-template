package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

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

func getEntriesFromDir(repo string, dir string, client *api.RESTClient, ch chan<- GitFile, wg *sync.WaitGroup) {
	defer wg.Done()
	response := []GitFile{}
	base := "repos/" + repo + "/contents/"
	err := client.Get(base+dir, &response)
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
	go getEntriesFromDir("github/gitignore", "", client, ch, &wg)

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
			go getEntriesFromDir("github/gitignore", file.Path, client, ch, &wg)
		}
	}
	return files, nil
}

func getGitUserName() (string, error) {
	cmd := exec.Command("git", "config", "user.name")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	name := strings.TrimSpace(string(out))
	if name == "" {
		return "", fmt.Errorf("git user.name is empty")
	}
	return name, nil
}

func getGitHubLogin(client *api.RESTClient) (string, error) {
	type viewer struct {
		Login string `json:"login"`
		Name  string `json:"name"`
	}
	var me viewer
	err := client.Get("user", &me)
	if err != nil {
		return "", err
	}
	// Prefer full profile name when available, fall back to login.
	if strings.TrimSpace(me.Name) != "" {
		return strings.TrimSpace(me.Name), nil
	}
	if strings.TrimSpace(me.Login) != "" {
		return strings.TrimSpace(me.Login), nil
	}
	return "", fmt.Errorf("github user identity is empty")
}

func templateLicenseBody(body string, fullName string) string {
	year := fmt.Sprintf("%d", time.Now().Year())
	result := strings.ReplaceAll(body, "[year]", year)
	if strings.TrimSpace(fullName) != "" {
		result = strings.ReplaceAll(result, "[fullname]", fullName)
	}
	return result
}

func main() {
	Cli := &cobra.Command{}

	IgnoreCmd := &cobra.Command{
		Use:     "gitignore",
		Aliases: []string{"ignore"},
		Example: "gh template gitignore --get Python",
		Short:   "list and download GitHub .gitignore templates",
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
			url := "repos/github/gitignore/contents/" + name
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

	LicenseCmd := &cobra.Command{
		Use:     "license",
		Aliases: []string{"licence", "licenses", "licences"},
		Example: "gh template license --get mit --save\ngh template license --list",
		Short:   "list and download license templates from GitHub",
	}
	licenseListFlag := LicenseCmd.Flags().BoolP("list", "l", false, "list available license templates")
	licenseGetFlag := LicenseCmd.Flags().StringP("get", "g", "", "get a license template by key (use `--list` to list available licenses)")
	licenseSaveFlag := LicenseCmd.Flags().BoolP("save", "s", false, "save downloaded license text to ./LICENSE")
	LicenseCmd.MarkFlagsMutuallyExclusive("get", "list")

	LicenseCmd.Run = func(cmd *cobra.Command, args []string) {
		client, err := api.DefaultRESTClient()
		if err != nil {
			fmt.Println(err)
			return
		}

		if *licenseListFlag {
			type LicenseListItem struct {
				Key  string
				Name string
			}
			items := []LicenseListItem{}
			err := client.Get("licenses", &items)
			if err != nil {
				fmt.Println(err)
				return
			}
			for _, item := range items {
				fmt.Printf("%s\t%s\n", item.Key, item.Name)
			}
			return
		}

		if cmd.Flags().Changed("get") {
			type LicenseDetail struct {
				Key  string
				Name string
				Body string
			}

			key := strings.ToLower(strings.TrimSpace(*licenseGetFlag))
			if key == "" {
				fmt.Println("license key cannot be empty")
				return
			}

			var detail LicenseDetail
			err := client.Get("licenses/"+key, &detail)
			if err != nil {
				fmt.Println(err)
				return
			}

			fullName, nameErr := getGitUserName()
			if nameErr != nil {
				fullName, nameErr = getGitHubLogin(client)
			}

			templated := templateLicenseBody(detail.Body, fullName)

			if strings.Contains(detail.Body, "[fullname]") {
				path := "./LICENSE"
				if nameErr != nil || strings.TrimSpace(fullName) == "" {
					fmt.Printf("Failed to template [fullname] in %s. You must fill it in yourself\n", path)
				} else {
					fmt.Printf("Templated [fullname] in license to %s\n", fullName)
				}
			}

			if *licenseSaveFlag {
				path := "LICENSE"
				err := os.WriteFile(path, []byte(templated), 0644)
				if err != nil {
					fmt.Println(err)
					return
				}
				fmt.Println("saved license to ./LICENSE")
				return
			}

			fmt.Printf("%s\n", templated)
		}
	}

	Cli.AddCommand(IgnoreCmd)
	Cli.AddCommand(LicenseCmd)
	err := Cli.Execute()
	if err != nil {
		fmt.Println(err)
	}
}