## GH-TEMPLATE

Download one of GitHub's excellent gitignore or license templates using the github cli tool.

## USAGE

### Gitignore templates

List available `.gitignore` templates:

gh template gitignore --list

Include community and global templates:

gh template gitignore --list --all

Print a specific template:

gh template gitignore --get Python

### License templates

List available licenses:

gh template license --list

Download and print a license template by key:

gh template license --get mit

Download and save to `./LICENSE` (GitHub default filename):

gh template license --get mit --save

## SOURCES

- [Templates](https://github.com/github/gitignore)
- [Licenses](https://api.github.com/licenses)
