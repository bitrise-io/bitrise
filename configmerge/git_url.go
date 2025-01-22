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

func parseGitRepoURL(gitURL string) (*GitRepoURL, error) {
	// https syntax: https://<host>[:<port>]/<path-to-git-repo>
	if strings.HasPrefix(gitURL, "https://") {
		u, err := url.Parse(gitURL)
		if err != nil {
			return nil, err
		}

		user := ""
		if u.User != nil {
			user = u.User.Username()
		}

		host := u.Hostname()
		port := u.Port()
		path := strings.TrimPrefix(u.Path, "/")

		return &GitRepoURL{
			User:           user,
			Host:           host,
			Port:           port,
			Path:           path,
			OriginalSyntax: HTTPSRepoURLSyntax,
		}, nil
	}

	// scp-like syntax: [<user>@]<host>:<path-to-git-repo>
	re := regexp.MustCompile(`^(?:(?P<user>[^@]+)@)?(?P<host>[^:]+):(?P<path>.+)$`)
	matches := re.FindStringSubmatch(gitURL)
	if matches == nil {
		return nil, fmt.Errorf("unsupported git URL format: %s", gitURL)
	}
	user := ""
	host := ""
	path := ""

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

	return &GitRepoURL{
		User:           user,
		Host:           host,
		Path:           path,
		OriginalSyntax: SSHGitRepoURLSyntax,
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
