package domain

type Novel struct {
	Title    string
	Content  string
	Chapters []Chapter
}

type Chapter struct {
	ID      string
	Title   string
	Order   int
	Content string
	Summary string
}

type Character struct {
	ID          string
	Name        string
	Role        string
	Description string
}

type Screenplay struct {
	SchemaVersion  string
	Title          string
	SourceType     string
	Language       string
	Provider       string
	Mode           string
	CreatedAt      string
	Characters     []Character
	SourceChapters []Chapter
	Acts           []Act
}

type Act struct {
	ID     string
	Title  string
	Order  int
	Scenes []Scene
}

type Scene struct {
	ID               string
	SourceChapterIDs []string
	Heading          Heading
	Summary          string
	Characters       []string
	Beats            []Beat
}

type Heading struct {
	Location string
	Time     string
	Interior bool
}

type Beat struct {
	Type          string
	Text          string
	CharacterID   string
	CharacterName string
}
