package botswizard

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	conversationVersionKey  = "_cv"
	conversationRevisionKey = "_cr"
	conversationFeatureKey  = "_cf"
	conversationFlowKey     = "_cfl"
	conversationStepKey     = "_cs"
	conversationExpiresKey  = "_ce"
	payloadPrefix           = "_cp_"
	payloadKeysKey          = "_cp_keys"

	maxConversationPayloadEntries  = 32
	maxConversationPayloadKeyLen   = 64
	maxConversationPayloadValueLen = 4096
)

var (
	ErrConversationExpired          = errors.New("conversation has expired")
	ErrConversationCorrupt          = errors.New("conversation state is corrupt")
	ErrConversationRevisionConflict = errors.New("conversation revision conflict")
)

// FeatureID identifies the feature which owns a conversation.
type FeatureID string

// FlowID identifies a flow within a feature.
type FlowID string

// StepID identifies the current step within a flow.
type StepID string

// ConversationState is the portable, versioned state contract for a feature
// flow. Payload is deliberately string-valued so it can be persisted through
// legacy BotChatData wizard params without a storage migration. Revision is
// observational when saved with Save; use SaveIfRevision to detect stale writes.
type ConversationState struct {
	Version   int
	Revision  int64
	Feature   FeatureID
	Flow      FlowID
	Step      StepID
	Payload   map[string]string
	ExpiresAt time.Time
}

func (v ConversationState) Expired(now time.Time) bool {
	return !v.ExpiresAt.IsZero() && !now.Before(v.ExpiresAt)
}

// ConversationAdapter migrates existing AwaitingReplyTo state lazily. New
// state is stored in namespaced wizard params; routing remains on the legacy
// AwaitingReplyTo path, preserving existing command matching behaviour.
type ConversationAdapter struct{ Now func() time.Time }

func (a ConversationAdapter) clock() time.Time {
	if a.Now != nil {
		return a.Now()
	}
	return time.Now()
}

// Save writes the supplied revision as-is. It is intended for migration and
// single-writer flows; concurrent callers must use SaveIfRevision.
func (a ConversationAdapter) Save(st state, value ConversationState) error {
	if err := validateConversation(value); err != nil {
		return err
	}
	if value.Version <= 0 {
		value.Version = 1
	}
	st.SetAwaitingReplyTo(string(value.Flow))
	st.AddWizardParam(conversationVersionKey, strconv.Itoa(value.Version))
	st.AddWizardParam(conversationRevisionKey, strconv.FormatInt(value.Revision, 10))
	st.AddWizardParam(conversationFeatureKey, string(value.Feature))
	st.AddWizardParam(conversationFlowKey, string(value.Flow))
	st.AddWizardParam(conversationStepKey, string(value.Step))
	if !value.ExpiresAt.IsZero() {
		st.AddWizardParam(conversationExpiresKey, value.ExpiresAt.UTC().Format(time.RFC3339))
	}
	keys := make([]string, 0, len(value.Payload))
	for key := range value.Payload {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		st.AddWizardParam(payloadPrefix+key, value.Payload[key])
	}
	if len(keys) > 0 {
		st.AddWizardParam(payloadKeysKey, strings.Join(keys, ","))
	}
	return nil
}

// SaveIfRevision stores the next revision only when expectedRevision matches
// the current state. An absent state has revision zero. This turns the legacy
// parameter store into a fail-closed optimistic-concurrency boundary.
func (a ConversationAdapter) SaveIfRevision(st state, value ConversationState, expectedRevision int64) error {
	if expectedRevision < 0 {
		return fmt.Errorf("%w: expected revision cannot be negative", ErrConversationRevisionConflict)
	}
	current, found, err := a.LoadChecked(st)
	if err != nil && !errors.Is(err, ErrConversationExpired) {
		return err
	}
	if found && current.Revision != expectedRevision {
		return fmt.Errorf("%w: have %d, expected %d", ErrConversationRevisionConflict, current.Revision, expectedRevision)
	}
	if !found && expectedRevision != 0 {
		return fmt.Errorf("%w: no state, expected %d", ErrConversationRevisionConflict, expectedRevision)
	}
	value.Revision = expectedRevision + 1
	return a.Save(st, value)
}

// Load preserves the original two-value API. It fails closed for expired or
// malformed persisted state; callers needing the reason should use LoadChecked.
func (a ConversationAdapter) Load(st state) (ConversationState, bool) {
	v, ok, _ := a.LoadChecked(st)
	return v, ok
}

// LoadChecked returns a typed error for malformed or expired persisted state.
// In either case it clears AwaitingReplyTo so a stale flow cannot be resumed.
func (a ConversationAdapter) LoadChecked(st state) (ConversationState, bool, error) {
	flow := st.GetWizardParam(conversationFlowKey)
	if flow == "" { // v0 migration: preserve existing AwaitingReplyTo as flow.
		flow = st.GetAwaitingReplyTo()
		if flow == "" {
			return ConversationState{}, false, nil
		}
		return ConversationState{Version: 0, Flow: FlowID(flow), Step: StepID(flow), Payload: map[string]string{}}, true, nil
	}
	version, err := parseNonNegativeInt(st.GetWizardParam(conversationVersionKey), 1)
	if err != nil {
		return a.corrupt(st, err)
	}
	revision, err := parseNonNegativeInt64(st.GetWizardParam(conversationRevisionKey), 0)
	if err != nil {
		return a.corrupt(st, err)
	}
	v := ConversationState{Version: version, Revision: revision, Feature: FeatureID(st.GetWizardParam(conversationFeatureKey)), Flow: FlowID(flow), Step: StepID(st.GetWizardParam(conversationStepKey)), Payload: map[string]string{}}
	if err := validateConversation(v); err != nil {
		return a.corrupt(st, err)
	}
	if expires := st.GetWizardParam(conversationExpiresKey); expires != "" {
		v.ExpiresAt, err = time.Parse(time.RFC3339, expires)
		if err != nil {
			return a.corrupt(st, err)
		}
		if v.Expired(a.clock()) {
			st.SetAwaitingReplyTo("")
			return ConversationState{}, false, ErrConversationExpired
		}
	}
	keysText := st.GetWizardParam(payloadKeysKey)
	if keysText != "" {
		keys := strings.Split(keysText, ",")
		if len(keys) > maxConversationPayloadEntries {
			return a.corrupt(st, errors.New("too many payload entries"))
		}
		for _, key := range keys {
			if err := validatePayloadEntry(key, st.GetWizardParam(payloadPrefix+key)); err != nil {
				return a.corrupt(st, err)
			}
			if _, duplicate := v.Payload[key]; duplicate {
				return a.corrupt(st, errors.New("duplicate payload key"))
			}
			v.Payload[key] = st.GetWizardParam(payloadPrefix + key)
		}
	}
	return v, true, nil
}

func (a ConversationAdapter) corrupt(st state, cause error) (ConversationState, bool, error) {
	st.SetAwaitingReplyTo("")
	return ConversationState{}, false, fmt.Errorf("%w: %v", ErrConversationCorrupt, cause)
}

func validateConversation(value ConversationState) error {
	if strings.TrimSpace(string(value.Feature)) == "" || strings.TrimSpace(string(value.Flow)) == "" {
		return errors.New("conversation feature and flow are required")
	}
	if value.Revision < 0 {
		return errors.New("conversation revision cannot be negative")
	}
	if len(value.Payload) > maxConversationPayloadEntries {
		return fmt.Errorf("conversation payload has more than %d entries", maxConversationPayloadEntries)
	}
	for key, payload := range value.Payload {
		if err := validatePayloadEntry(key, payload); err != nil {
			return err
		}
	}
	return nil
}

func validatePayloadEntry(key, value string) error {
	if key == "" || len(key) > maxConversationPayloadKeyLen || strings.Contains(key, ",") {
		return errors.New("conversation payload key is invalid")
	}
	if len(value) > maxConversationPayloadValueLen {
		return errors.New("conversation payload value is too large")
	}
	return nil
}

func parseNonNegativeInt(raw string, defaultValue int) (int, error) {
	if raw == "" {
		return defaultValue, nil
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v < 0 {
		return 0, errors.New("invalid non-negative integer")
	}
	return v, nil
}

func parseNonNegativeInt64(raw string, defaultValue int64) (int64, error) {
	if raw == "" {
		return defaultValue, nil
	}
	v, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || v < 0 {
		return 0, errors.New("invalid non-negative int64")
	}
	return v, nil
}

func (a ConversationAdapter) Cancel(st state) { st.SetAwaitingReplyTo("") }
func IsUniversalCancel(text string) bool {
	return strings.EqualFold(strings.TrimSpace(text), "/cancel") || strings.EqualFold(strings.TrimSpace(text), "cancel")
}
