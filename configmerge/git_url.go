package configmerge

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

type GitRepoURL struct {
	User string
	Host string
	Port string
	Path string
}

func isHttpFormatRepoURL(gitURL string) bool {
	return strings.HasPrefix(gitURL, "http://") || strings.HasPrefix(gitURL, "https://")
}

func parseGitRepoURL(gitURL string) (*GitRepoURL, error) {
	if strings.HasPrefix(gitURL, "ssh://") || // ssh syntax: ssh://[<user>@]<host>[:<port>]/<path-to-git-repo>
		strings.HasPrefix(gitURL, "git://") || // git syntax: git://<host>[:<port>]/<path-to-git-repo>
		strings.HasPrefix(gitURL, "http://") || strings.HasPrefix(gitURL, "https://") || // http[s] syntax: http[s]://<host>[:<port>]/<path-to-git-repo>
		strings.HasPrefix(gitURL, "ftp://") || strings.HasPrefix(gitURL, "ftps://") { // ftp[s] syntax: ftp[s]://<host>[:<port>]/<path-to-git-repo>
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
			User: user,
			Host: host,
			Port: port,
			Path: path,
		}, nil
	} else { // SCP-like syntax: [<user>@]<host>:/<path-to-git-repo>
		re := regexp.MustCompile(`^(?:(?P<user>[^@]+)@)?(?P<host>[^:]+):(?P<path>.+)$`)
		matches := re.FindStringSubmatch(gitURL)
		if matches != nil {
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
				User: user,
				Host: host,
				Path: path,
			}, nil
		}
	}

	return nil, fmt.Errorf("unsupported git URL format")
}

func generateSCPStyleSSHFormatRepoURL(details *GitRepoURL) string {
	var urlBuilder strings.Builder

	// SSH format: [<user>@]<host>:/<path-to-git-repo>
	if details.User != "" {
		urlBuilder.WriteString(details.User)
		urlBuilder.WriteString("@")
	}
	urlBuilder.WriteString(details.Host)
	urlBuilder.WriteString(":/")
	urlBuilder.WriteString(details.Path)

	return urlBuilder.String()
}

func equalGitRepoURLs(a, b *GitRepoURL) bool {
	return a.Host == b.Host && a.Path == b.Path
}
