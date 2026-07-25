package botswizard

import (
	"errors"
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
)

// ConversationState is the portable, versioned state contract for a feature
// flow. Payload is deliberately string-valued so it can be persisted through
// legacy BotChatData wizard params without a storage migration.
type ConversationState struct {
	Version   int
	Revision  int64
	Feature   string
	Flow      string
	Step      string
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

func (a ConversationAdapter) Save(st state, value ConversationState) error {
	if strings.TrimSpace(value.Feature) == "" || strings.TrimSpace(value.Flow) == "" {
		return errors.New("conversation feature and flow are required")
	}
	if value.Version <= 0 {
		value.Version = 1
	}
	if value.Revision < 0 {
		return errors.New("conversation revision cannot be negative")
	}
	st.SetAwaitingReplyTo(value.Flow)
	st.AddWizardParam(conversationVersionKey, strconv.Itoa(value.Version))
	st.AddWizardParam(conversationRevisionKey, strconv.FormatInt(value.Revision, 10))
	st.AddWizardParam(conversationFeatureKey, value.Feature)
	st.AddWizardParam(conversationFlowKey, value.Flow)
	st.AddWizardParam(conversationStepKey, value.Step)
	if !value.ExpiresAt.IsZero() {
		st.AddWizardParam(conversationExpiresKey, value.ExpiresAt.UTC().Format(time.RFC3339))
	}
	keys := make([]string, 0, len(value.Payload))
	for key, payload := range value.Payload {
		if key == "" || strings.Contains(key, ",") {
			return errors.New("conversation payload keys must be non-empty and cannot contain commas")
		}
		st.AddWizardParam(payloadPrefix+key, payload)
		keys = append(keys, key)
	}
	if len(keys) > 0 {
		st.AddWizardParam(payloadKeysKey, strings.Join(keys, ","))
	}
	return nil
}

func (a ConversationAdapter) Load(st state) (ConversationState, bool) {
	flow := st.GetWizardParam(conversationFlowKey)
	if flow == "" { // v0 migration: preserve existing AwaitingReplyTo as flow.
		flow = st.GetAwaitingReplyTo()
		if flow == "" {
			return ConversationState{}, false
		}
		return ConversationState{Version: 0, Flow: flow, Step: flow, Payload: map[string]string{}}, true
	}
	v := ConversationState{Version: atoiDefault(st.GetWizardParam(conversationVersionKey), 1), Revision: int64(atoiDefault(st.GetWizardParam(conversationRevisionKey), 0)), Feature: st.GetWizardParam(conversationFeatureKey), Flow: flow, Step: st.GetWizardParam(conversationStepKey), Payload: map[string]string{}}
	if expires := st.GetWizardParam(conversationExpiresKey); expires != "" {
		v.ExpiresAt, _ = time.Parse(time.RFC3339, expires)
	}
	for _, key := range strings.Split(st.GetWizardParam(payloadKeysKey), ",") {
		if key != "" {
			v.Payload[key] = st.GetWizardParam(payloadPrefix + key)
		}
	}
	return v, true
}

func (a ConversationAdapter) Cancel(st state) { st.SetAwaitingReplyTo("") }
func IsUniversalCancel(text string) bool {
	return strings.EqualFold(strings.TrimSpace(text), "/cancel") || strings.EqualFold(strings.TrimSpace(text), "cancel")
}
