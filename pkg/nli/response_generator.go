package nli

import (
	"context"
	"fmt"
	"strings"
)

// ============================================================================
// Response Generator
// ============================================================================

// DefaultResponseGenerator generates responses based on intent and query results
type DefaultResponseGenerator struct {
	config    Config
	templates map[Intent][]string
}

// NewDefaultResponseGenerator creates a new response generator
func NewDefaultResponseGenerator(config Config) *DefaultResponseGenerator {
	gen := &DefaultResponseGenerator{
		config:    config,
		templates: make(map[Intent][]string),
	}
	gen.initializeTemplates()
	return gen
}

// initializeTemplates sets up response templates for each intent
func (g *DefaultResponseGenerator) initializeTemplates() {
	g.templates[IntentQueryBalance] = []string{
		"Your current balance is %s.",
		"You have %s in your wallet.",
		"Your wallet contains %s.",
	}

	g.templates[IntentFindOfferings] = []string{
		"I found %d offerings matching your criteria:\n%s",
		"Here are the available offerings:\n%s",
		"Based on your search, here are some options:\n%s",
	}

	g.templates[IntentGetHelp] = []string{
		"I'd be happy to help! %s",
		"Here's what you need to know: %s",
		"Great question! %s",
	}

	g.templates[IntentCheckOrder] = []string{
		"Your order status: %s",
		"Order %s is currently %s.",
		"Here's the status of your order:\n%s",
	}

	g.templates[IntentGetProviderInfo] = []string{
		"Here's information about provider %s:\n%s",
		"Provider details:\n%s",
	}

	g.templates[IntentStaking] = []string{
		"Here's what you need to know about staking: %s",
		"Regarding staking: %s",
	}

	g.templates[IntentDeployment] = []string{
		"Here's how to work with deployments: %s",
		"Regarding your deployment: %s",
	}

	g.templates[IntentIdentity] = []string{
		"Your identity verification status: %s",
		"Here's your VEID information: %s",
	}

	g.templates[IntentGeneralChat] = []string{
		"%s",
		"I understand. %s",
	}
}

// Generate generates a response for the given intent and context
func (g *DefaultResponseGenerator) Generate(ctx context.Context, intent Intent, entities map[string]string, queryResult *QueryResult, chatCtx *ChatContext) (string, error) {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	switch intent {
	case IntentQueryBalance:
		return g.generateBalanceResponse(queryResult, entities)
	case IntentFindOfferings:
		return g.generateOfferingsResponse(queryResult, entities)
	case IntentGetHelp:
		return g.generateHelpResponse(entities)
	case IntentCheckOrder:
		return g.generateOrderResponse(queryResult, entities)
	case IntentGetProviderInfo:
		return g.generateProviderResponse(queryResult, entities)
	case IntentStaking:
		return g.generateStakingResponse(queryResult, entities)
	case IntentDeployment:
		return g.generateDeploymentResponse(queryResult, entities)
	case IntentIdentity:
		return g.generateIdentityResponse(queryResult, entities, chatCtx)
	case IntentGeneralChat:
		return g.generateGeneralResponse(entities)
	default:
		return g.generateGeneralResponse(entities)
	}
}

// generateBalanceResponse generates a balance query response
//
//nolint:unparam // entities kept for future address-specific response formatting
func (g *DefaultResponseGenerator) generateBalanceResponse(result *QueryResult, _ map[string]string) (string, error) {
	if result == nil || !result.Success {
		return "I couldn't retrieve your balance at this time. Please try again later or check if your address is connected.", nil
	}

	balances, ok := result.Data.([]BalanceInfo)
	if !ok {
		return "Your balance information is available, but I encountered an issue formatting it.", nil
	}

	if len(balances) == 0 {
		return "Your wallet appears to be empty. You don't have any tokens yet.", nil
	}

	var sb strings.Builder
	sb.WriteString("Here's your current balance:\n")
	for _, b := range balances {
		sb.WriteString(fmt.Sprintf("• %s %s", b.Amount, strings.ToUpper(b.Denom)))
		if b.USD != "" {
			sb.WriteString(fmt.Sprintf(" (≈ $%s)", b.USD))
		}
		sb.WriteString("\n")
	}

	return sb.String(), nil
}

// generateOfferingsResponse generates a marketplace offerings response
func (g *DefaultResponseGenerator) generateOfferingsResponse(result *QueryResult, entities map[string]string) (string, error) {
	if result == nil || !result.Success {
		return "I couldn't search the marketplace right now. Please try again later.", nil
	}

	offerings, ok := result.Data.([]OfferingInfo)
	if !ok {
		return "I found some offerings, but encountered an issue displaying them.", nil
	}

	if len(offerings) == 0 {
		resourceType := entities["resource_type"]
		if resourceType != "" {
			return fmt.Sprintf("I couldn't find any %s offerings matching your criteria. Try broadening your search.", resourceType), nil
		}
		return "No offerings match your search criteria. Try a different search.", nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("I found %d offerings:\n\n", len(offerings)))

	for i, o := range offerings {
		if i >= 5 {
			sb.WriteString(fmt.Sprintf("\n... and %d more offerings.", len(offerings)-5))
			break
		}
		sb.WriteString(fmt.Sprintf("**%d. %s**\n", i+1, o.Type))
		sb.WriteString(fmt.Sprintf("   Provider: %s\n", truncateAddress(o.Provider)))
		sb.WriteString(fmt.Sprintf("   Price: %s\n", o.Price))
		if o.Available {
			sb.WriteString("   Status: ✅ Available\n")
		} else {
			sb.WriteString("   Status: ⚠️ Currently unavailable\n")
		}
		sb.WriteString("\n")
	}

	return sb.String(), nil
}

// generateHelpResponse generates a help response
func (g *DefaultResponseGenerator) generateHelpResponse(entities map[string]string) (string, error) {
	// This would typically be enhanced with LLM for contextual help
	return `Here are some things I can help you with:

• **Balance**: "What's my balance?" or "Show my tokens"
• **Marketplace**: "Find GPU providers" or "Search for compute offerings"
• **Orders**: "Check my order status" or "Show my deployments"
• **Staking**: "How do I stake tokens?" or "Show staking rewards"
• **VEID**: "What's my identity score?" or "How do I verify my identity?"
• **Deployments**: "How do I deploy an app?" or "Create a deployment"

Just ask me anything about VirtEngine!`, nil
}

// generateOrderResponse generates an order status response
//
//nolint:unparam // entities kept for future order ID formatting
func (g *DefaultResponseGenerator) generateOrderResponse(result *QueryResult, _ map[string]string) (string, error) {
	if result == nil || !result.Success {
		return "I couldn't retrieve your order information. Please check if you have the correct order ID.", nil
	}

	orders, ok := result.Data.([]OrderInfo)
	if !ok {
		order, okSingle := result.Data.(OrderInfo)
		if okSingle {
			orders = []OrderInfo{order}
		} else {
			return "I found your order information, but couldn't format it properly.", nil
		}
	}

	if len(orders) == 0 {
		return "I couldn't find any orders matching your query.", nil
	}

	var sb strings.Builder
	if len(orders) == 1 {
		o := orders[0]
		sb.WriteString(fmt.Sprintf("**Order %s**\n", o.ID))
		sb.WriteString(fmt.Sprintf("Status: %s\n", formatStatus(o.Status)))
		sb.WriteString(fmt.Sprintf("Provider: %s\n", truncateAddress(o.Provider)))
		sb.WriteString(fmt.Sprintf("Created: %s\n", o.CreatedAt.Format("Jan 2, 2006")))
	} else {
		sb.WriteString(fmt.Sprintf("You have %d orders:\n\n", len(orders)))
		for _, o := range orders {
			sb.WriteString(fmt.Sprintf("• **%s**: %s (Provider: %s)\n", o.ID, formatStatus(o.Status), truncateAddress(o.Provider)))
		}
	}

	return sb.String(), nil
}

// generateProviderResponse generates a provider info response
func (g *DefaultResponseGenerator) generateProviderResponse(result *QueryResult, entities map[string]string) (string, error) {
	if result == nil || !result.Success {
		address := entities["address"]
		if address != "" {
			return fmt.Sprintf("I couldn't find information about provider %s.", truncateAddress(address)), nil
		}
		return "I couldn't find provider information. Please specify a provider address.", nil
	}

	return fmt.Sprintf("Provider information:\n%v", result.Data), nil
}

// generateStakingResponse generates a staking-related response
func (g *DefaultResponseGenerator) generateStakingResponse(result *QueryResult, entities map[string]string) (string, error) {
	return `**Staking on VirtEngine**

To stake your tokens:
1. Go to the Staking section in the portal
2. Choose a validator to delegate to
3. Enter the amount you want to stake
4. Confirm the transaction

**Key points:**
• Minimum stake: 1 UVE
• Unbonding period: 21 days
• Rewards are distributed automatically
• You can redelegate to a different validator at any time

Need help choosing a validator? Ask me about "validator recommendations".`, nil
}

// generateDeploymentResponse generates a deployment-related response
func (g *DefaultResponseGenerator) generateDeploymentResponse(result *QueryResult, entities map[string]string) (string, error) {
	return `**Creating a Deployment on VirtEngine**

To deploy an application:
1. Prepare your SDL (Stack Definition Language) manifest
2. Go to the Deployments section
3. Upload or paste your SDL
4. Choose provider(s) from the bids
5. Fund and start the deployment

**Example SDL structure:**
` + "```yaml\nversion: \"2.0\"\nservices:\n  web:\n    image: nginx:latest\n    expose:\n      - port: 80\n        as: 80\n        to:\n          - global: true\n```" + `

Need help with a specific deployment configuration? Just ask!`, nil
}

// generateIdentityResponse generates an identity/VEID response
//
//nolint:unparam // result kept for future identity data extraction
func (g *DefaultResponseGenerator) generateIdentityResponse(_ *QueryResult, _ map[string]string, chatCtx *ChatContext) (string, error) {
	if chatCtx != nil && chatCtx.UserProfile != nil {
		profile := chatCtx.UserProfile
		if profile.IsVerified {
			return fmt.Sprintf(`**Your VEID Status: ✅ Verified**

Identity Score: %.1f%%
Verified Address: %s

Your account is fully verified and you have access to all VirtEngine features.`,
				profile.IdentityScore*100, truncateAddress(profile.Address)), nil
		}
		return `**Your VEID Status: ⏳ Not Verified**

To verify your identity:
1. Go to the VEID section in the portal
2. Complete the identity verification flow
3. Upload required documents
4. Complete the liveness check

Verification enables full marketplace access and higher transaction limits.`, nil
	}

	return `**About VEID (VirtEngine Identity)**

VEID is VirtEngine's decentralized identity verification system. It uses ML-powered scoring to verify user identities while preserving privacy.

**Benefits of verification:**
• Full marketplace access
• Higher transaction limits
• Provider trust signals
• Reduced friction on sensitive operations

To check your verification status, connect your wallet first.`, nil
}

// generateGeneralResponse generates a general chat response
func (g *DefaultResponseGenerator) generateGeneralResponse(entities map[string]string) (string, error) {
	return `I'm here to help you with VirtEngine! You can ask me about:
• Your token balance and transactions
• Marketplace offerings and providers
• Order status and deployments
• Staking and delegation
• Identity verification (VEID)

What would you like to know?`, nil
}

// ============================================================================
// Helper Functions
// ============================================================================

// truncateAddress truncates a blockchain address for display
func truncateAddress(address string) string {
	if len(address) <= 16 {
		return address
	}
	return address[:8] + "..." + address[len(address)-6:]
}

// formatStatus formats an order status for display
func formatStatus(status string) string {
	switch strings.ToLower(status) {
	case "open":
		return "🔵 Open"
	case "active":
		return "🟢 Active"
	case "closed":
		return "⚪ Closed"
	case "matched":
		return "🟡 Matched"
	case "failed":
		return "🔴 Failed"
	default:
		return status
	}
}

