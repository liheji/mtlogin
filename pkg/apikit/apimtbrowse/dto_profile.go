package apimtbrowse

// ProfileData maps the /api/member/profile response data.
type ProfileData struct {
	Username     string              `json:"username"`
	MemberStatus ProfileMemberStatus `json:"memberStatus"`
	MemberCount  ProfileMemberCount  `json:"memberCount"`
}

// ProfileMemberStatus maps the memberStatus sub-object.
type ProfileMemberStatus struct {
	LastLogin  string `json:"lastLogin"`
	LastBrowse string `json:"lastBrowse"`
}

// ProfileMemberCount maps the memberCount sub-object.
type ProfileMemberCount struct {
	Bonus      float64 `json:"bonus,string"`
	Uploaded   int64   `json:"uploaded,string"`
	Downloaded int64   `json:"downloaded,string"`
}
