package keeper

import (
	"sync"

	"github.com/virtengine/virtengine/x/veid/types"
)

type developmentMLScorerSlot struct {
	mu     sync.Mutex
	scorer MLScorer
}

type lockedDevelopmentMLScorer struct {
	slot *developmentMLScorerSlot
}

func newDevelopmentMLScorerSlot() *developmentMLScorerSlot {
	return &developmentMLScorerSlot{}
}

func (s *developmentMLScorerSlot) set(scorer MLScorer) {
	_ = s.replace(scorer, false)
}

func (s *developmentMLScorerSlot) replace(scorer MLScorer, closePrevious bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if wrapsDevelopmentMLScorerSlot(scorer, s) {
		return types.ErrMLInferenceFailed.Wrap("cannot install development scorer slot wrapper")
	}
	if closePrevious && s.scorer != nil && s.scorer != scorer {
		if wrapsDevelopmentMLScorerSlot(s.scorer, s) {
			s.scorer = nil
			return nil
		}
		if err := s.scorer.Close(); err != nil {
			return err
		}
	}
	s.scorer = scorer
	return nil
}

func (s *developmentMLScorerSlot) closeActive() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.scorer == nil {
		return nil
	}
	if wrapsDevelopmentMLScorerSlot(s.scorer, s) {
		s.scorer = nil
		return nil
	}
	err := s.scorer.Close()
	s.scorer = nil
	return err
}

func (s *developmentMLScorerSlot) scorerOrFailClosed() MLScorer {
	if s == nil {
		return &failClosedMLScorer{
			err: types.ErrMLInferenceFailed.Wrap("no production inference scorer is configured; use signed receipts"),
		}
	}
	return &lockedDevelopmentMLScorer{slot: s}
}

func (s *lockedDevelopmentMLScorer) Score(input *ScoringInput) (*ScoringOutput, error) {
	s.slot.mu.Lock()
	defer s.slot.mu.Unlock()
	if s.slot.scorer == nil {
		return nil, types.ErrMLInferenceFailed.Wrap("no production inference scorer is configured; use signed receipts")
	}
	return s.slot.scorer.Score(input)
}

func (s *lockedDevelopmentMLScorer) GetModelVersion() string {
	s.slot.mu.Lock()
	defer s.slot.mu.Unlock()
	if s.slot.scorer == nil {
		return ""
	}
	return s.slot.scorer.GetModelVersion()
}

func (s *lockedDevelopmentMLScorer) IsHealthy() bool {
	s.slot.mu.Lock()
	defer s.slot.mu.Unlock()
	return s.slot.scorer != nil && s.slot.scorer.IsHealthy()
}

func (s *lockedDevelopmentMLScorer) Close() error {
	return s.slot.closeActive()
}

func wrapsDevelopmentMLScorerSlot(scorer MLScorer, slot *developmentMLScorerSlot) bool {
	wrapped, ok := scorer.(*lockedDevelopmentMLScorer)
	return ok && wrapped.slot == slot
}
