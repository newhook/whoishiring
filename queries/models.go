// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0

package queries

type Embedding struct {
	ID        int    `json:"id"`
	Model     string `json:"model"`
	ItemID    int    `json:"item_id"`
	Embedding []byte `json:"embedding"`
	CreatedAt int    `json:"created_at"`
	UpdatedAt int    `json:"updated_at"`
}

type Item struct {
	ID          int    `json:"id"`
	Deleted     bool   `json:"deleted"`
	Type        string `json:"type"`
	By          string `json:"by"`
	Time        int    `json:"time"`
	Text        string `json:"text"`
	Dead        bool   `json:"dead"`
	Parent      int    `json:"parent"`
	Poll        int    `json:"poll"`
	Url         string `json:"url"`
	Score       int    `json:"score"`
	Title       string `json:"title"`
	Descendants int    `json:"descendants"`
}

type ItemKid struct {
	ItemID int `json:"item_id"`
	KidID  int `json:"kid_id"`
}

type ItemPart struct {
	ItemID int `json:"item_id"`
	PartID int `json:"part_id"`
}