package database

const TmplStrOnUse = `package tbc

import (
{{- if .HasStacking }}
	"time"

{{ end }}
	"github.com/wowsims/tbc/sim/common/shared"
{{- if .HasStacking }}
	"github.com/wowsims/tbc/sim/core"
{{- end }}
)

func RegisterAllOnUseCds() {
{{- range .Groups }}

	// {{ .Name }}
{{- range .Entries }}
	{{- if .Skipped}}
	{{- range (.Tooltip | formatStrings 100) }}
	// Not simulated: {{.}}
	{{- end}}
	{{with index .Variants 0 -}}
	// https://www.wowhead.com/tbc/spell={{.SpellID}}
	{{- end}}
	{{- else}}
	{{- if .StackingOnUse}}
  	{{- with index .Variants 0}}
	// {{ .Name }} - https://www.wowhead.com/tbc/spell={{.SpellID}}
	{{- end}}
	shared.NewStackingStatBonusCD(shared.StackingStatBonusCD{
		Name:                  "{{ .StackingOnUse.Name }}",
		ID:                    {{ (index .Variants 0).ID }},
		CD:                    time.Millisecond * {{ .StackingOnUse.CooldownMs }},
		Callback:              {{ .StackProcInfo.Callback | asCoreCallback }},
		ProcMask:              {{ .StackProcInfo.ProcMask | asCoreProcMask }},
		Outcome:               {{ .StackProcInfo.Outcome | asCoreOutcome }},
		RequireDamageDealt:    {{ .StackProcInfo.RequireDamageDealt }},
		TrinketLimitsDuration: true,
	})
	{{- else if not .Supported}}
  	{{- with index .Variants 0}}
	// shared.NewSimpleStatActive({{ .ID }}) // {{ .Name }} - https://www.wowhead.com/tbc/spell={{.SpellID}}
	{{- end}}
	{{- else}}
  	{{- with index .Variants 0}}
	shared.NewSimpleStatActive({{ .ID }}) // {{ .Name }} - https://www.wowhead.com/tbc/spell={{.SpellID}}
	{{- end}}
	{{- end}}
	{{- end}}
{{- end }}

{{- end }}
}`
const TmplStrProc = `package tbc

import (
{{- if .HasDamageIcd }}
	"time"

{{ end }}
	"github.com/wowsims/tbc/sim/core"
 	"github.com/wowsims/tbc/sim/common/shared"
)

func RegisterAllProcs() {
{{- range .Groups }}

	// {{ .Name }}
{{- range .Entries }}
	{{- if .Skipped}}
	{{- range (.Tooltip | formatStrings 100) }}
	// Not simulated: {{.}}
	{{- end}}
	{{with index .Variants 0 -}}
	// https://www.wowhead.com/tbc/spell={{.SpellID}}
	{{- end}}
	{{- else}}
	{{if not .Supported}}
	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	{{- end}}
	{{- range (.Tooltip | formatStrings 100) }}
	// {{.}}
	{{- end}}
	{{with index .Variants 0 -}}
	// https://www.wowhead.com/tbc/spell={{.SpellID}}
	{{- end}}
	{{- if .Supported}}
		{{- if .Damage}}
			{{- $entry := . }}
			{{- range .Variants }}
			shared.NewProcDamageEffect(shared.ProcDamageEffect{
				ItemID:  {{ .ID }},
				SpellID: {{ $entry.Damage.SpellID }},
				School:  {{ $entry.Damage.SchoolMask | asCoreSpellSchool }},
				MinDmg:  {{ $entry.Damage.MinDamage }},
				MaxDmg:  {{ $entry.Damage.MaxDamage }},
				{{- if $entry.DamageCannotCrit }}
				CannotCrit: true,
				{{- end}}
				Flags:   core.SpellFlagNoOnCastComplete | core.SpellFlagPassiveSpell | core.SpellFlagNoOnDamageDealt,
				Trigger: core.ProcTrigger{
					Name:               "{{ .Name }}",
					ActionID:           core.ActionID{ItemID: {{ .ID }}},
					Callback:           {{ $entry.ProcInfo.Callback | asCoreCallback }},
					ProcMask:           {{ $entry.ProcInfo.ProcMask | asCoreProcMask }},
					Outcome:            {{ $entry.ProcInfo.Outcome | asCoreOutcome }},
					RequireDamageDealt: {{ $entry.ProcInfo.RequireDamageDealt }},
					ProcChance:         {{ $entry.DamageProcChance }},
					{{- if $entry.DamageIcdMs }}
					ICD:                time.Millisecond * {{ $entry.DamageIcdMs }},
					{{- end}}
				},
			})
			{{- end}}
		{{- else if gt .ProcInfo.MaxCumulativeStacks 0 }}
			shared.NewStackingStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
				Callback:           {{ .ProcInfo.Callback | asCoreCallback }},
				ProcMask:           {{ .ProcInfo.ProcMask | asCoreProcMask }},
				Outcome:            {{ .ProcInfo.Outcome | asCoreOutcome }},
				RequireDamageDealt: {{ .ProcInfo.RequireDamageDealt }},
				{{- if .ProcInfo.ClassSpellsOnly }}
				ClassSpellsOnly:    {{ .ProcInfo.ClassSpellsOnly }},
				{{- end}}
			{{- if .StackProcInfo }}
				StackCallback:      {{ .StackProcInfo.Callback | asCoreCallback }},
				StackProcMask:      {{ .StackProcInfo.ProcMask | asCoreProcMask }},
				StackOutcome:       {{ .StackProcInfo.Outcome | asCoreOutcome }},
			{{- end}}
			}, []shared.ItemVariant{
				{{- range .Variants }}
				{ItemID: {{.ID}}, ItemName: "{{.Name}}"},
				{{- end}}
			})
		{{- else}}
			shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
				Callback:           {{ .ProcInfo.Callback | asCoreCallback }},
				ProcMask:           {{ .ProcInfo.ProcMask | asCoreProcMask }},
				Outcome:            {{ .ProcInfo.Outcome | asCoreOutcome }},
				RequireDamageDealt: {{ .ProcInfo.RequireDamageDealt }},
				{{- if .ProcInfo.ClassSpellsOnly }}
				ClassSpellsOnly:    {{ .ProcInfo.ClassSpellsOnly }},
				{{- end}}
			{{- if .StackProcInfo }}
				StackCallback:      {{ .StackProcInfo.Callback | asCoreCallback }},
				StackProcMask:      {{ .StackProcInfo.ProcMask | asCoreProcMask }},
				StackOutcome:       {{ .StackProcInfo.Outcome | asCoreOutcome }},
			{{- end}}
			}, []shared.ItemVariant{
				{{- range .Variants }}
				{ItemID: {{.ID}}, ItemName: "{{.Name}}"},
				{{- end}}
			})
		{{- end}}
	{{- else}}
		{{- if gt .ProcInfo.MaxCumulativeStacks 0 }}
			// shared.NewStackingStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
			//	Callback:           {{ .ProcInfo.Callback | asCoreCallback }},
			//	ProcMask:           {{ .ProcInfo.ProcMask | asCoreProcMask }},
			//	Outcome:            {{ .ProcInfo.Outcome | asCoreOutcome }},
			//	RequireDamageDealt: {{ .ProcInfo.RequireDamageDealt }},
			{{- if .ProcInfo.ClassSpellsOnly }}
			//	ClassSpellsOnly:    {{ .ProcInfo.ClassSpellsOnly }},
			{{- end}}
			// }, []shared.ItemVariant{
				{{- range .Variants }}
			//	{ItemID: {{.ID}}, ItemName: "{{.Name}}"},
				{{- end}}
			// })
		{{- else}}
			// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
			//	Callback:           {{ .ProcInfo.Callback | asCoreCallback }},
			//	ProcMask:           {{ .ProcInfo.ProcMask | asCoreProcMask }},
			//	Outcome:            {{ .ProcInfo.Outcome | asCoreOutcome }},
			//	RequireDamageDealt: {{ .ProcInfo.RequireDamageDealt }}
			{{- if .ProcInfo.ClassSpellsOnly }}
			// ClassSpellsOnly:    {{ .ProcInfo.ClassSpellsOnly }},
			{{- end}}
			// }, []shared.ItemVariant{
				{{- range .Variants }}
			//	{ItemID: {{.ID}}, ItemName: "{{.Name}}"},
				{{- end}}
			// })
		{{- end}}
	{{- end}}
	{{- end}}
{{- end }}

{{- end }}
}`

const TmplStrEnchant = `package tbc
{{ if .HasEntries }}
import (
	"github.com/wowsims/tbc/sim/core"
 	"github.com/wowsims/tbc/sim/common/shared"
)
{{- end }}

func RegisterAllEnchants() {
{{- range .Groups }}

	// {{ .Name }}
{{- range .Entries }}
	{{- if .Skipped}}
	{{- range (.Tooltip | formatStrings 100) }}
	// Not simulated: {{.}}
	{{- end}}
	{{with index .Variants 0 -}}
	// https://www.wowhead.com/tbc/spell={{.SpellID}}
	{{- end}}
	{{- else}}
	{{if not .Supported}}
	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	{{- end}}
	{{- range (.Tooltip | formatStrings 100) }}
	// {{.}}
	{{- end}}
	{{with index .Variants 0 -}}
	// https://www.wowhead.com/tbc/spell={{.SpellID}}
	{{- end}}
	{{- if .Supported}}
		shared.NewProcStatBonusEffect(shared.ProcStatBonusEffect{
			{{with index .Variants 0 -}}
			Name:               "{{ .Name }}",
			EnchantID:          {{ .ID }},
			{{- end}}
			Callback:           {{ .ProcInfo.Callback | asCoreCallback }},
			ProcMask:           {{ .ProcInfo.ProcMask | asCoreProcMask }},
			Outcome:            {{ .ProcInfo.Outcome | asCoreOutcome }},
			RequireDamageDealt: {{ .ProcInfo.RequireDamageDealt }},
			{{- if .ProcInfo.ClassSpellsOnly }}
			ClassSpellsOnly:    {{ .ProcInfo.ClassSpellsOnly }},
			{{- end}}
		})
	{{- else}}
		// shared.NewProcStatBonusEffect(shared.ProcStatBonusEffect{
		{{- with index .Variants 0 }}
		//	Name:               "{{ .Name }}",
		//	EnchantID:          {{ .ID }},
		{{- end}}
		//	Callback:           {{ .ProcInfo.Callback | asCoreCallback }},
		//	ProcMask:           {{ .ProcInfo.ProcMask | asCoreProcMask }},
		//	Outcome:            {{ .ProcInfo.Outcome | asCoreOutcome }},
		//	RequireDamageDealt: {{ .ProcInfo.RequireDamageDealt }},
		{{- if .ProcInfo.ClassSpellsOnly }}
		//  ClassSpellsOnly:    {{ .ProcInfo.ClassSpellsOnly }},
		{{- end}}
		// })
	{{- end}}
	{{- end}}
{{- end }}

{{- end }}
}`

const TmplStrMissingEffects = `
// This file is auto generated
// Changes will be overwritten on next database generation

export const MISSING_ITEM_EFFECTS = new Map<number, string[]>([
{{- range .ItemEffects }}
	[
		{{.ItemID}}, // {{ .Name }}
		[
			{{- range .Effects }}
			"{{ jsString .Name }}", // {{.SpellID}} - https://www.wowhead.com/tbc/spell={{.SpellID}}
			{{- end}}
		]
	],
{{- end }}
])

export const MISSING_ENCHANT_EFFECTS = new Map<number, string[]>([
{{- range .EnchantEffects }}
	[
		{{.ItemID}}, // {{ .Name }}
		[
			{{- range .Effects }}
			"{{ jsString .Name }}", // {{.SpellID}} - https://www.wowhead.com/tbc/spell={{.SpellID}}
			{{- end}}
		]
	],
{{- end }}
])
`
