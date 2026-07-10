package scenario

import "github.com/solguardlabs/northstardtl/src/domain"

type TemplateOption func(*Definition)

func NewQuoteTemplate(name string, intents ...domain.Intent) Definition {
	actions := make([]Action, 0, len(intents)+1)
	for index, intent := range intents {
		actions = append(actions, Action{
			Type:   "quote",
			Label:  labelWithIndex("quote", index),
			Intent: intent,
		})
	}
	actions = append(actions, Action{Type: "snapshot", Label: "snapshot"})
	return Definition{Name: name, Bootstrap: DefaultBootstrap(), Actions: actions}
}

func NewSettlementTemplate(name string, intent domain.Intent, options ...TemplateOption) Definition {
	definition := Definition{
		Name:      name,
		Bootstrap: DefaultBootstrap(),
		Actions: []Action{
			{Type: "submit", Label: "admit", Intent: intent},
			{Type: "advance_epoch", Label: "ready", Delta: 1},
			{Type: "execute", Label: "execute", Count: 1},
			{Type: "snapshot", Label: "snapshot"},
		},
	}
	for _, option := range options {
		option(&definition)
	}
	return definition
}

func WithBootstrap(bootstrap BootstrapAlias) TemplateOption {
	return func(definition *Definition) {
		definition.Bootstrap = bootstrap.ToAPI()
	}
}

func WithActionBeforeExecute(action Action) TemplateOption {
	return func(definition *Definition) {
		if len(definition.Actions) < 3 {
			definition.Actions = append(definition.Actions, action)
			return
		}
		updated := make([]Action, 0, len(definition.Actions)+1)
		updated = append(updated, definition.Actions[:2]...)
		updated = append(updated, action)
		updated = append(updated, definition.Actions[2:]...)
		definition.Actions = updated
	}
}

func labelWithIndex(prefix string, index int) string {
	if index == 0 {
		return prefix
	}
	return prefix + "-" + string(rune('a'+index))
}
