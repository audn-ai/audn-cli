package validation

import (
	"fmt"
	"regexp"
	"strings"
)

// UUID regex pattern (standard UUID format)
var uuidRegex = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// ValidateUUID checks if a string is a valid UUID format
func ValidateUUID(uuid string) error {
	uuid = strings.TrimSpace(uuid)
	if uuid == "" {
		return fmt.Errorf("UUID cannot be empty")
	}

	if !uuidRegex.MatchString(uuid) {
		return fmt.Errorf("invalid UUID format: %s", uuid)
	}

	return nil
}

// ValidateAgentID validates an agent ID (must be UUID)
func ValidateAgentID(agentID string) error {
	if err := ValidateUUID(agentID); err != nil {
		return fmt.Errorf("invalid agent ID: %w", err)
	}
	return nil
}

// ValidateCampaignID validates a campaign ID (must be UUID)
func ValidateCampaignID(campaignID string) error {
	if err := ValidateUUID(campaignID); err != nil {
		return fmt.Errorf("invalid campaign ID: %w", err)
	}
	return nil
}
