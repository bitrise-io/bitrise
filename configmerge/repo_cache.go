package configmerge

type repoCache struct {
	cache map[string]string
}

func NewRepoCache() RepoCache {
	return repoCache{
		cache: map[string]string{},
	}
}

func (c repoCache) GetRepo(ref ConfigReference) string {
	return c.cache[ref.RepoKey()]
}

func (c repoCache) SetRepo(dir string, ref ConfigReference) {
	c.cache[ref.RepoKey()] = dir
}
