package model

import "time"

// AIKnowledge AI 客服知识库条目；BreedID 为空表示平台通用知识
type AIKnowledge struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	BreedID   *int64    `json:"breedId"`
	Title     string    `json:"title"`
	Keywords  string    `json:"keywords"`
	Content   string    `json:"content"`
	Sort      int       `json:"sort"`
	Status    int       `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (AIKnowledge) TableName() string { return "ai_knowledge" }
