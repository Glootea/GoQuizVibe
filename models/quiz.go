package models

import (
	"encoding/json"

	"github.com/goquizvibe/db"
)

const (
	QuestionTypeChoice = db.QuestionTypeChoice
	QuestionTypeOpen   = db.QuestionTypeOpen
	QuestionTypeFill   = db.QuestionTypeFill
)

type TextSegment struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

func ParseFillSegments(correctAnswer string) ([]TextSegment, error) {
	if correctAnswer == "" {
		return []TextSegment{}, nil
	}
	var segments []TextSegment
	if err := json.Unmarshal([]byte(correctAnswer), &segments); err != nil {
		return nil, err
	}
	return segments, nil
}

func MarshalFillSegments(segments []TextSegment) (string, error) {
	data, err := json.Marshal(segments)
	if err != nil {
		return "", err
	}
	return string(data), nil
}