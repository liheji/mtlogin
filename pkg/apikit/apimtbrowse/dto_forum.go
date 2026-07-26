package apimtbrowse

// ForumsData maps the /api/forum/forums response data.
type ForumsData struct {
	ForumsList []ForumsEntry                  `json:"forumsList"`
	LastPost   map[string]ForumsLastPostEntry `json:"lastPost"`
}

// ForumsEntry maps a single forum list entry.
type ForumsEntry struct {
	ID int64 `json:"id,string"`
}

// ForumsLastPostEntry maps a single lastPost entry keyed by forum ID.
type ForumsLastPostEntry struct {
	Post ForumsPost `json:"post"`
}

// ForumsPost maps a post within a lastPost entry.
type ForumsPost struct {
	AuthorID int64 `json:"authorId,string"`
}

// TopicSearchData maps the /api/forum/topic/search response data.
type TopicSearchData struct {
	Data []TopicSearchEntry `json:"data"`
}

// TopicSearchEntry maps a single topic in search results.
type TopicSearchEntry struct {
	ID       int64               `json:"id,string"`
	Author   int64               `json:"author,string"`
	LastPost TopicSearchLastPost `json:"lastPost"`
}

// TopicSearchLastPost maps the lastPost within a topic search entry.
type TopicSearchLastPost struct {
	AuthorID int64 `json:"authorId,string"`
}
