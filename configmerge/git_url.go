package configmerge

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

type GitRepoURLSyntax string

const (
	SSHGitRepoURLSyntax GitRepoURLSyntax = "ssh"
	HTTPSRepoURLSyntax  GitRepoURLSyntax = "https"
)

type GitRepoURL struct {
	User           string
	Host           string
	Port           string
	Path           string
	OriginalSyntax GitRepoURLSyntax
}

func NewGitRepoURL(gitURL string) (*GitRepoURL, error) {
	var user, host, port, path string
	var syntax GitRepoURLSyntax

	// https syntax: https://<host>[:<port>]/<path-to-git-repo>
	if strings.HasPrefix(gitURL, "https://") {
		u, err := url.Parse(gitURL)
		if err != nil {
			return nil, err
		}

		if u.User != nil {
			user = u.User.Username()
		}

		host = u.Hostname()
		port = u.Port()
		path = strings.TrimPrefix(u.Path, "/")
		syntax = HTTPSRepoURLSyntax
	} else {
		// scp-like syntax: [<user>@]<host>:<path-to-git-repo>
		re := regexp.MustCompile(`^(?:(?P<user>[^@]+)@)?(?P<host>[^:]+):(?P<path>.+)$`)
		matches := re.FindStringSubmatch(gitURL)
		if matches == nil {
			return nil, fmt.Errorf("unsupported git URL format: %s", gitURL)
		}

		for i, name := range re.SubexpNames() {
			switch name {
			case "user":
				user = matches[i]
			case "host":
				host = matches[i]
			case "path":
				path = matches[i]
			}
		}

		syntax = SSHGitRepoURLSyntax
	}

	pathComponents := strings.Split(path, "/")
	if len(pathComponents) < 2 {
		return nil, fmt.Errorf("repository path (%s) is expected in a 'user/repo_name' format", path)
	}

	return &GitRepoURL{
		User:           user,
		Host:           host,
		Port:           port,
		Path:           path,
		OriginalSyntax: syntax,
	}, nil
}

func (u GitRepoURL) URLString(syntax GitRepoURLSyntax) string {
	var urlBuilder strings.Builder

	if syntax == HTTPSRepoURLSyntax {
		// https syntax: http[s]://<host>[:<port>]/<path-to-git-repo>
		urlBuilder.WriteString("https://")
		urlBuilder.WriteString(u.Host)
		if u.Port != "" {
			urlBuilder.WriteString(":")
			urlBuilder.WriteString(u.Port)
		}
		urlBuilder.WriteString("/")
		urlBuilder.WriteString(u.Path)
	} else {
		// scp-like syntax: [<user>@]<host>:<path-to-git-repo>
		if u.User != "" {
			urlBuilder.WriteString(u.User)
			urlBuilder.WriteString("@")
		}
		urlBuilder.WriteString(u.Host)
		urlBuilder.WriteString(":")
		urlBuilder.WriteString(u.Path)
	}

	return urlBuilder.String()
}

func (u GitRepoURL) RepoURLForRepo(repoName string) GitRepoURL {
	if repoName == "" {
		return GitRepoURL{
			User:           u.User,
			Host:           u.Host,
			Port:           u.Port,
			Path:           u.Path,
			OriginalSyntax: u.OriginalSyntax,
		}
	}

	var path string
	pathComponents := strings.Split(u.Path, "/")
	if len(pathComponents) < 2 {
		path = repoName + ".git"
	} else {
		path = strings.Join(pathComponents[:len(pathComponents)-1], "/") + "/" + repoName + ".git"
	}

	return GitRepoURL{
		User:           u.User,
		Host:           u.Host,
		Port:           u.Port,
		Path:           path,
		OriginalSyntax: u.OriginalSyntax,
	}
}
