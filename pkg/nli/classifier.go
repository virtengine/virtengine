package nli

import (
	"context"
	"regexp"
	"strings"
)

// ============================================================================
// Intent Classifier
// ============================================================================

// RuleBasedClassifier implements intent classification using pattern matching.
// This classifier uses keyword patterns and rules to classify intents without
// requiring an LLM, making it fast and deterministic.
type RuleBasedClassifier struct {
	// patterns maps intents to their keyword patterns
	patterns map[Intent][]PatternRule
}

// PatternRule defines a pattern matching rule
type PatternRule struct {
	// Pattern is the regex pattern to match
	Pattern *regexp.Regexp

	// Keywords are simple keyword matches
	Keywords []string

	// Weight is the confidence weight for this pattern
	Weight float32

	// RequiredEntities are entity types that must be extractable
	RequiredEntities []string
}

// NewRuleBasedClassifier creates a new rule-based classifier
func NewRuleBasedClassifier() *RuleBasedClassifier {
	c := &RuleBasedClassifier{
		patterns: make(map[Intent][]PatternRule),
	}
	c.initializePatterns()
	return c
}

// initializePatterns sets up the default pattern rules
func (c *RuleBasedClassifier) initializePatterns() {
	// Balance queries
	c.patterns[IntentQueryBalance] = []PatternRule{
		{
			Keywords: []string{"balance", "tokens", "coins", "funds", "wallet"},
			Weight:   0.8,
		},
		{
			Pattern: regexp.MustCompile(`(?i)(how much|what('s| is)|show|check).*(balance|tokens?|uve|money|funds)`),
			Weight:  0.9,
		},
		{
			Pattern: regexp.MustCompile(`(?i)my (tokens?|balance|funds|coins)`),
			Weight:  0.85,
		},
	}

	// Marketplace offerings search
	c.patterns[IntentFindOfferings] = []PatternRule{
		{
			Keywords: []string{"offerings", "marketplace", "providers", "compute", "gpu", "storage"},
			Weight:   0.75,
		},
		{
			Pattern: regexp.MustCompile(`(?i)(find|search|show|list|browse).*(offerings?|providers?|compute|gpu|storage|resources?)`),
			Weight:  0.9,
		},
		{
			Pattern: regexp.MustCompile(`(?i)(available|cheap|best).*(providers?|offerings?|gpu|compute)`),
			Weight:  0.85,
		},
		{
			Pattern: regexp.MustCompile(`(?i)looking for.*(compute|gpu|storage|server)`),
			Weight:  0.85,
		},
	}

	// Help requests
	c.patterns[IntentGetHelp] = []PatternRule{
		{
			Keywords: []string{"help", "how to", "how do", "explain", "what is", "tutorial"},
			Weight:   0.7,
		},
		{
			Pattern: regexp.MustCompile(`(?i)(how (do|can|to)|help|explain|teach|show me how)`),
			Weight:  0.8,
		},
		{
			Pattern: regexp.MustCompile(`(?i)what (is|are|does) (a |the )?(veid|virtengine|staking|deployment)`),
			Weight:  0.85,
		},
	}

	// Order status
	c.patterns[IntentCheckOrder] = []PatternRule{
		{
			Keywords: []string{"order", "status", "lease", "deployment status"},
			Weight:   0.75,
		},
		{
			Pattern: regexp.MustCompile(`(?i)(check|show|track|what('s| is)).*(order|lease|deployment|status)`),
			Weight:  0.9,
		},
		{
			Pattern: regexp.MustCompile(`(?i)my (orders?|leases?|deployments?)`),
			Weight:  0.85,
		},
	}

	// Provider information
	c.patterns[IntentGetProviderInfo] = []PatternRule{
		{
			Keywords: []string{"provider info", "about provider", "provider details"},
			Weight:   0.8,
		},
		{
			Pattern: regexp.MustCompile(`(?i)(tell|info|details|more).*(about|on) (provider|ve1[a-z0-9]+)`),
			Weight:  0.9,
		},
		{
			Pattern: regexp.MustCompile(`(?i)provider (ve1[a-z0-9]+|[a-zA-Z]+)`),
			Weight:  0.8,
		},
	}

	// Staking
	c.patterns[IntentStaking] = []PatternRule{
		{
			Keywords: []string{"stake", "staking", "delegate", "delegation", "validator", "unstake", "redelegate"},
			Weight:   0.85,
		},
		{
			Pattern: regexp.MustCompile(`(?i)(how to |can i )?(stake|delegate|unstake|redelegate|bond|unbond)`),
			Weight:  0.9,
		},
		{
			Pattern: regexp.MustCompile(`(?i)(staking|delegation) rewards?`),
			Weight:  0.85,
		},
	}

	// Deployment
	c.patterns[IntentDeployment] = []PatternRule{
		{
			Keywords: []string{"deploy", "deployment", "manifest", "SDL", "workload"},
			Weight:   0.8,
		},
		{
			Pattern: regexp.MustCompile(`(?i)(create|start|launch|make).*(deployment|workload|container|app)`),
			Weight:  0.9,
		},
		{
			Pattern: regexp.MustCompile(`(?i)(deploy|run).*(app|container|service|workload)`),
			Weight:  0.9,
		},
	}

	// Identity/VEID
	c.patterns[IntentIdentity] = []PatternRule{
		{
			Keywords: []string{"veid", "identity", "verification", "kyc", "verify", "verified"},
			Weight:   0.85,
		},
		{
			Pattern: regexp.MustCompile(`(?i)(verify|verified|verification|identity) (my |account |status)?`),
			Weight:  0.9,
		},
		{
			Pattern: regexp.MustCompile(`(?i)(veid|identity) score`),
			Weight:  0.95,
		},
	}
}

// Classify classifies the intent of a message
func (c *RuleBasedClassifier) Classify(ctx context.Context, message string, chatCtx *ChatContext) (*ClassificationResult, error) {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	message = strings.TrimSpace(message)
	lowerMessage := strings.ToLower(message)

	var bestIntent Intent = IntentGeneralChat
	var bestConfidence float32 = 0.0
	var alternatives []IntentScore

	// Score each intent
	for intent, rules := range c.patterns {
		score := c.scoreIntent(lowerMessage, rules)
		if score > 0 {
			alternatives = append(alternatives, IntentScore{Intent: intent, Confidence: score})
			if score > bestConfidence {
				bestConfidence = score
				bestIntent = intent
			}
		}
	}

	// Extract entities
	entities := c.extractEntities(message)

	// If no strong match, default to general chat with base confidence
	if bestConfidence < 0.3 {
		bestIntent = IntentGeneralChat
		bestConfidence = 0.5
	}

	// Sort alternatives by confidence (simple bubble sort for small array)
	for i := range alternatives {
		for j := i + 1; j < len(alternatives); j++ {
			if alternatives[j].Confidence > alternatives[i].Confidence {
				alternatives[i], alternatives[j] = alternatives[j], alternatives[i]
			}
		}
	}

	// Keep only top 3 alternatives
	if len(alternatives) > 3 {
		alternatives = alternatives[:3]
	}

	return &ClassificationResult{
		Intent:             bestIntent,
		Confidence:         bestConfidence,
		Entities:           entities,
		AlternativeIntents: alternatives,
	}, nil
}

// scoreIntent calculates the confidence score for an intent
func (c *RuleBasedClassifier) scoreIntent(message string, rules []PatternRule) float32 {
	var maxScore float32 = 0.0

	for _, rule := range rules {
		var score float32 = 0.0

		// Check regex pattern
		if rule.Pattern != nil && rule.Pattern.MatchString(message) {
			score = rule.Weight
		}

		// Check keywords
		if score == 0 && len(rule.Keywords) > 0 {
			matchCount := 0
			for _, keyword := range rule.Keywords {
				if strings.Contains(message, strings.ToLower(keyword)) {
					matchCount++
				}
			}
			if matchCount > 0 {
				// Score based on percentage of keywords matched
				keywordScore := float32(matchCount) / float32(len(rule.Keywords))
				score = keywordScore * rule.Weight
			}
		}

		if score > maxScore {
			maxScore = score
		}
	}

	return maxScore
}

// extractEntities extracts entities from the message
func (c *RuleBasedClassifier) extractEntities(message string) map[string]string {
	entities := make(map[string]string)

	// Extract blockchain addresses (ve1...)
	addressPattern := regexp.MustCompile(`ve1[a-z0-9]{38,}`)
	if matches := addressPattern.FindStringSubmatch(message); len(matches) > 0 {
		entities["address"] = matches[0]
	}

	// Extract order/deployment IDs (numeric patterns)
	idPattern := regexp.MustCompile(`(?i)(order|deployment|lease)\s*#?\s*(\d+)`)
	if matches := idPattern.FindStringSubmatch(message); len(matches) > 2 {
		entities["order_id"] = matches[2]
	}

	// Extract token amounts
	amountPattern := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*(uve|tokens?|uact)`)
	if matches := amountPattern.FindStringSubmatch(message); len(matches) > 1 {
		entities["amount"] = matches[1]
		if len(matches) > 2 {
			entities["denom"] = matches[2]
		}
	}

	// Extract resource types
	resourcePattern := regexp.MustCompile(`(?i)(gpu|cpu|storage|memory|ram|compute)`)
	if matches := resourcePattern.FindStringSubmatch(message); len(matches) > 0 {
		entities["resource_type"] = strings.ToLower(matches[1])
	}

	return entities
}

// ============================================================================
// LLM-Based Classifier
// ============================================================================

// LLMClassifier uses an LLM backend for intent classification
type LLMClassifier struct {
	backend LLMBackend
}

// NewLLMClassifier creates a new LLM-based classifier
func NewLLMClassifier(backend LLMBackend) *LLMClassifier {
	return &LLMClassifier{backend: backend}
}

// Classify uses the LLM to classify intent
func (c *LLMClassifier) Classify(ctx context.Context, message string, chatCtx *ChatContext) (*ClassificationResult, error) {
	if c.backend == nil {
		return nil, ErrLLMBackendNotConfigured
	}
	return c.backend.ClassifyIntent(ctx, message)
}

// ============================================================================
// Hybrid Classifier
// ============================================================================

// HybridClassifier combines rule-based and LLM classification
type HybridClassifier struct {
	ruleClassifier *RuleBasedClassifier
	llmClassifier  *LLMClassifier
	threshold      float32
}

// NewHybridClassifier creates a hybrid classifier
func NewHybridClassifier(backend LLMBackend, threshold float32) *HybridClassifier {
	return &HybridClassifier{
		ruleClassifier: NewRuleBasedClassifier(),
		llmClassifier:  NewLLMClassifier(backend),
		threshold:      threshold,
	}
}

// Classify first tries rule-based classification, falls back to LLM if confidence is low
func (c *HybridClassifier) Classify(ctx context.Context, message string, chatCtx *ChatContext) (*ClassificationResult, error) {
	// Try rule-based classification first
	result, err := c.ruleClassifier.Classify(ctx, message, chatCtx)
	if err != nil {
		return nil, err
	}

	// If confidence is high enough, return rule-based result
	if result.Confidence >= c.threshold {
		return result, nil
	}

	// Fall back to LLM classification
	if c.llmClassifier.backend != nil {
		llmResult, err := c.llmClassifier.Classify(ctx, message, chatCtx)
		if err == nil && llmResult.Confidence > result.Confidence {
			// Merge entities from rule-based extraction
			for k, v := range result.Entities {
				if _, exists := llmResult.Entities[k]; !exists {
					llmResult.Entities[k] = v
				}
			}
			return llmResult, nil
		}
	}

	return result, nil
}
