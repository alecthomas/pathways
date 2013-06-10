package main

import (
	"github.com/alecthomas/pathways"
)

type ScoreService struct {
}

func (s *ScoreService) GetScores(cx *pathways.Context) *pathways.Response {
	return nil
}

func (s *ScoreService) CreateScore(cx *pathways.Context) pathways.Response {
	return nil
}

func AuthFilter(cx *pathways.Context) pathways.StageAction {
	return pathways.StageCancel
}

func main() {
	scores := &ScoreService{}

	service := pathways.NewService()
	service.Path("/scores/").Get().ApiMethod(scores, "GetScores")
	service.Path("/scores/").Filter(AuthFilter).Post().ApiMethod(scores, "CreateScore")
}
