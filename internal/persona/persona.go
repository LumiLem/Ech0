package persona

import (
	"math/rand"
	"strings"
)

const (
	PersonaKey = "persona" // Persona 上下文及存储键
)

// Dimension 人格特征维度
type Dimension string

const (
	StyleDim      Dimension = "style"      // 风格维度
	MoodDim       Dimension = "mood"       // 情绪维度
	TopicsDim     Dimension = "topics"     // 兴趣主题维度
	ExpressionDim Dimension = "expression" // 表达偏好维度
)

// Feature 特征及其权重
type Feature struct {
	Name   string  `json:"name"`   // 特征名称
	Weight float64 `json:"weight"` // 特征权重 (0~1)
}

// Persona 平行人格描述
type Persona struct {
	ID           uint      `gorm:"primaryKey"                       json:"id"`           // 人格ID
	Name         string    `gorm:"type:varchar(50);unique;not null" json:"name"`         // 平行人格名称
	Style        []Feature `gorm:"type:text"                        json:"style"`        // JSON格式存储风格维度权重
	Mood         []Feature `gorm:"type:text"                        json:"mood"`         // JSON格式存储情绪维度权重
	Topics       []Feature `gorm:"type:text"                        json:"topics"`       // JSON格式存储兴趣主题权重
	Expression   []Feature `gorm:"type:text"                        json:"expression"`   // JSON格式存储表达偏好权重
	Description  string    `gorm:"type:text"                        json:"description"`  // 人格描述
	Independence float64   `gorm:"type:float"                       json:"independence"` // 独立性权重 (0~1)
	LastActive   int64     `                                        json:"last_active"`  // 上次生成内容时间
	CreatedAt    int64     `                                        json:"created_at"`
	UpdatedAt    int64     `                                        json:"updated_at"`
}

func (p *Persona) UpdateStyle(style []Feature) {
	p.Style = style
}

func (p *Persona) UpdateMood(mood []Feature) {
	p.Mood = mood
}

func (p *Persona) UpdateTopics(topics []Feature) {
	p.Topics = topics
}

func (p *Persona) UpdateExpression(expression []Feature) {
	p.Expression = expression
}

func (p *Persona) UpdateDescription(description string) {
	p.Description = description
}

// WhatDimensionToUpdate 随机选择一个维度进行更新
func (p *Persona) WhatDimensionToUpdate() Dimension {
	// 动态根据 feature 数量做反向加权，越少越容易被选中
	weights := map[Dimension]float64{
		StyleDim:      1.0 / float64(len(p.Style)+1),
		MoodDim:       1.0 / float64(len(p.Mood)+1),
		TopicsDim:     1.0 / float64(len(p.Topics)+1),
		ExpressionDim: 1.0 / float64(len(p.Expression)+1),
	}

	// 求总和
	var total float64
	for _, w := range weights {
		total += w
	}

	// 做一个 0~total 的 roulette wheel（轮盘赌随机）
	r := rand.Float64() * total

	for dim, w := range weights {
		if r < w {
			return dim
		}
		r -= w
	}

	return StyleDim // 理论不会走到这里
}

// GetDimensionFeatures 获取指定维度的特征
func (p *Persona) GetDimensionFeatures(dim Dimension) []Feature {
	switch dim {
	case StyleDim:
		return p.Style
	case MoodDim:
		return p.Mood
	case TopicsDim:
		return p.Topics
	case ExpressionDim:
		return p.Expression
	default:
		return nil
	}
}

// UpdateDimension 根据维度更新对应的特征
func (p *Persona) UpdateDimension(dim Dimension, features []Feature) {
	features = sanitizeFeatures(features)

	switch dim {
	case StyleDim:
		p.UpdateStyle(features)
	case MoodDim:
		p.UpdateMood(features)
	case TopicsDim:
		p.UpdateTopics(features)
	case ExpressionDim:
		p.UpdateExpression(features)
	}
}

// sanitizeFeatures 清理和规范化特征列表
func sanitizeFeatures(features []Feature) []Feature {
	clean := make([]Feature, 0, len(features))

	for _, f := range features {
		// name 必须存在且非空
		if strings.TrimSpace(f.Name) == "" {
			continue
		}

		// weight 修正到 0~1
		if f.Weight < 0 {
			f.Weight = 0
		}
		if f.Weight > 1 {
			f.Weight = 1
		}

		clean = append(clean, f)
	}

	return clean
}
