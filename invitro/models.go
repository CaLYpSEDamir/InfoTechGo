package invitro

// SearchHelper структура для полных описаний исследований
type SearchHelper struct {
	Link_88 string
	Link_84 string
	Link_81 string
	Link_82 string
}

// ResearchType модель типов исследований
type ResearchType struct {
	ID       uint `gorm:"primary_key"`
	ParentID uint
	Name     string `gorm:"size:255"`
}

// Research модель исследований
type Research struct {
	ID          uint   `gorm:"primary_key"`
	Name        string `gorm:"size:500"`
	TypeID      uint
	Description string
	Training    string
	Indication  string
	Result      string
}
