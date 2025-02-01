package todoist

// Part of paid version
type (
	Duration interface{}
	Deadline interface{}
)

type Project struct {
	Id             *string `json:"id"`
	Name           *string `json:"name"`
	Color          *string `json:"color"`
	ParentId       *string `json:"parent_id"`
	Order          int     `json:"order"`
	CommentCount   int     `json:"comment_count"`
	IsShared       bool    `json:"is_shared"`
	IsFavorite     bool    `json:"is_favorite"`
	IsInboxProject bool    `json:"is_inbox_project"`
	IsTeamInbox    bool    `json:"is_team_inbox"`
	ViewStyle      *string `json:"view_style"`
	Url            *string `json:"url"`
}

type Task struct {
	Id           *string   `json:"id"`
	ProjectId    *string   `json:"project_id"`
	SectionId    *string   `json:"section_id"`
	Content      *string   `json:"content"`
	Description  *string   `json:"description"`
	IsCompleted  bool      `json:"is_completed"`
	Labels       []string  `json:"labels"`
	ParentId     *string   `json:"parent_id"`
	Order        int       `json:"order"`
	Priority     int       `json:"priority"`
	Due          *Due      `json:"due"`
	Deadline     *Deadline `json:"deadline"`
	Url          *string   `json:"url"`
	CommentCount int       `json:"comment_count"`
	CreatedAt    *string   `json:"created_at"`
	CreatorId    *string   `json:"creator_id"`
	AssigneeId   *string   `json:"assignee_id"`
	AssignerId   *string   `json:"assigner_id"`
	Duration     *Duration `json:"duration"`
}

type Due struct {
	Name        *string `json:"string"`
	Date        *string `json:"date"`
	IsRecurring bool    `json:"is_recurring"`
	Datetime    *string `json:"datetime"`
	Timezone    *string `json:"timezone"`
	Lang        *string `json:"lang"`
}
