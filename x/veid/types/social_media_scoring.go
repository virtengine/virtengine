package types

import "time"

// SocialMediaScoringWeight defines the base weight for a provider.
type SocialMediaScoringWeight struct {
	Provider SocialMediaProviderType `json:"provider"`
	Weight   uint32                  `json:"weight"`
}

const (
	socialMediaVerifiedBonus  = 120
	socialMediaNameMatchBonus = 150
)

// DefaultSocialMediaScoringWeights returns default base weights for providers.
func DefaultSocialMediaScoringWeights() []SocialMediaScoringWeight {
	return []SocialMediaScoringWeight{
		{Provider: SocialMediaProviderGoogle, Weight: 280},
		{Provider: SocialMediaProviderFacebook, Weight: 220},
		{Provider: SocialMediaProviderMicrosoft, Weight: 300},
	}
}

// GetSocialMediaScoringWeight returns the base scoring weight for a provider.
func GetSocialMediaScoringWeight(provider SocialMediaProviderType) uint32 {
	for _, w := range DefaultSocialMediaScoringWeights() {
		if w.Provider == provider {
			return w.Weight
		}
	}
	return 0
}

// SocialMediaAccountAgeBonus returns a bonus based on account age.
func SocialMediaAccountAgeBonus(ageDays uint32) uint32 {
	switch {
	case ageDays >= 1460:
		return 220
	case ageDays >= 730:
		return 170
	case ageDays >= 365:
		return 130
	case ageDays >= 180:
		return 90
	case ageDays >= 90:
		return 60
	default:
		return 0
	}
}

// CalculateSocialMediaScore calculates the score contribution for a social media scope.
func CalculateSocialMediaScore(scope *SocialMediaScope, nameMatch bool, now time.Time) uint32 {
	if scope == nil {
		return 0
	}

	ageDays := scope.AccountAgeDays
	if ageDays == 0 && scope.AccountCreatedAt != nil && !scope.AccountCreatedAt.IsZero() {
		if now.After(*scope.AccountCreatedAt) {
			ageDays = uint32(now.Sub(*scope.AccountCreatedAt).Hours() / 24)
		}
	}

	score := GetSocialMediaScoringWeight(scope.Provider)
	score += SocialMediaAccountAgeBonus(ageDays)

	if scope.IsVerified {
		score += socialMediaVerifiedBonus
	}
	if nameMatch {
		score += socialMediaNameMatchBonus
	}

	return score
}
