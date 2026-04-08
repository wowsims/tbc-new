import tippy from 'tippy.js';
import { ref } from 'tsx-vanilla';

import { IndividualSimUI } from '../../individual_sim_ui';
import i18n from '../../../i18n/config';
import { translatePresetConfigurationCategory } from '../../../i18n/localization';
import { PresetBuild } from '../../preset_utils';
import { ConsumesSpec, Debuffs, Encounter, EquipmentSpec, HealingModel, IndividualBuffs, ItemSwap, PartyBuffs, RaidBuffs, Spec } from '../../proto/common';
import { SavedTalents } from '../../proto/ui';
import { isEqualAPLRotation } from '../../proto_utils/apl_utils';
import { Stats } from '../../proto_utils/stats';
import { TypedEvent } from '../../typed_event';
import { Component } from '../component';
import { ContentBlock } from '../content_block';
import { PresetGroupPicker, PresetGroupItem } from './preset_group_picker';

export enum PresetConfigurationCategory {
	EPWeights = 'epWeights',
	Gear = 'gear',
	Talents = 'talents',
	Rotation = 'rotation',
	RotationType = 'rotationType',
	Encounter = 'encounter',
	Settings = 'settings',
}

export class PresetConfigurationPicker extends Component {
	readonly simUI: IndividualSimUI<Spec>;
	readonly builds: Array<PresetBuild>;
	private readonly types?: PresetConfigurationCategory[];

	constructor(parentElem: HTMLElement, simUI: IndividualSimUI<Spec>, types?: PresetConfigurationCategory[]) {
		super(parentElem, 'preset-configuration-picker-root');
		this.rootElem.classList.add('saved-data-manager-root');

		this.simUI = simUI;
		this.types = types;
		this.builds = (this.simUI.individualConfig.presets.builds ?? []).filter(build =>
			Object.keys(build).some(category => types?.includes(category as PresetConfigurationCategory) && !!build[category as PresetConfigurationCategory]),
		);

		if (!this.builds.length) {
			this.rootElem.classList.add('hide');
			return;
		}

		const hasGrouping = this.builds.some(b => b.phase !== undefined || b.group !== undefined);
		if (hasGrouping) {
			this.buildGrouped();
		} else {
			this.buildFlat();
		}
	}

	private buildGrouped() {
		const groupPicker = new PresetGroupPicker(this.rootElem, {
			storageKey: this.simUI.getPresetFilterStorageKey(),
		});

		this.simUI.sim.waitForInit().then(() => {
			const buildItems: PresetGroupItem[] = this.builds.map(build => ({
				phase: build.phase,
				group: build.group,
				elem: this.makeBuildChip(build, groupPicker),
			}));
			groupPicker.addSection(i18n.t('gear_tab.preset_configurations.title'), buildItems);
			groupPicker.init();
		});
	}

	private buildFlat() {
		const contentBlock = new ContentBlock(this.rootElem, 'saved-data', {
			header: {
				title: i18n.t('gear_tab.preset_configurations.title'),
				tooltip: i18n.t('gear_tab.preset_configurations.tooltip'),
			},
		});

		const buildsContainerRef = ref<HTMLDivElement>();

		const container = (
			<div className="saved-data-container">
				<div className="saved-data-presets" ref={buildsContainerRef}></div>
			</div>
		);

		this.simUI.sim.waitForInit().then(() => {
			this.builds.forEach(build => {
				buildsContainerRef.value!.appendChild(this.makeBuildChip(build));
			});
			contentBlock.bodyElement.replaceChildren(container);
		});
	}

	private makeBuildChip(build: PresetBuild, groupPicker?: PresetGroupPicker): HTMLElement {
		const dataElemRef = ref<HTMLButtonElement>();
		const chip = (
			<button className="saved-data-set-chip badge rounded-pill" ref={dataElemRef}>
				<span
					className="saved-data-set-name"
					attributes={{ role: 'button' }}
					onclick={() => {
						PresetConfigurationPicker.applyBuild(TypedEvent.nextEventID(), build, this.simUI);
						groupPicker?.setFilter(build.phase);
					}}>
					{build.name}
				</span>
			</button>
		);

		let categories: string[] = [];

		// Add main categories from build keys
		Object.keys(build).forEach(c => {
			if (!['name', 'encounter', 'settings'].includes(c) && build[c as PresetConfigurationCategory]) {
				const category = c as PresetConfigurationCategory;
				categories.push(translatePresetConfigurationCategory(category));
			}
		});

		if (build.encounter?.encounter) {
			categories.push(translatePresetConfigurationCategory(PresetConfigurationCategory.Encounter));
		}

		if (build.epWeights) {
			categories.push(i18n.t('common.preset.stat_weights'));
		}

		if (build.settings) {
			Object.keys(build.settings).forEach(c => {
				if (c === 'name') {
					return;
				} else if (c === 'specOptions') {
					categories.push(i18n.t('common.preset.class_spec_options'));
				} else if (c === 'consumables') {
					categories.push(i18n.t('common.preset.consumables'));
				} else if (c === 'reforgeSettings') {
					categories.push(i18n.t('common.preset.reforge_settings'));
				} else if (c === 'debuffs') {
					categories.push(i18n.t('common.preset.debuffs'));
				} else if (['buffs', 'raidBuffs', 'partyBuffs'].includes(c)) {
					categories.push(i18n.t('common.preset.buffs'));
				} else {
					categories.push(i18n.t('common.preset.other_settings'));
				}
			});
		}

		categories = [...new Set(categories)].sort();

		tippy(dataElemRef.value!, {
			content: (
				<>
					<p className="mb-1">{i18n.t('common.preset.description')}</p>
					<ul className="mb-0">
						{categories.map(category => (
							<li>{category}</li>
						))}
					</ul>
				</>
			),
		});

		let rafId = 0;
		const checkActive = () => {
			cancelAnimationFrame(rafId);
			rafId = requestAnimationFrame(() => {
				const el = dataElemRef.value!;
				if (!el.isConnected) return;
				el.classList[this.isBuildActive(build) ? 'add' : 'remove']('active');
			});
		};

		checkActive();
		TypedEvent.onAny([
			this.simUI.player.changeEmitter,
			this.simUI.sim.settingsChangeEmitter,
			this.simUI.sim.raid.changeEmitter,
			this.simUI.sim.encounter.changeEmitter,
		]).on(checkActive);
		groupPicker?.onFilter(checkActive);

		return chip as HTMLElement;
	}

	static applyBuild(
		eventID: number,
		{ gear, itemSwap, rotation, rotationType, talents, epWeights, encounter, settings }: PresetBuild,
		simUI: IndividualSimUI<any>,
	) {
		TypedEvent.freezeAllAndDo(() => {
			if (gear) simUI.player.setGear(eventID, simUI.sim.db.lookupEquipmentSpec(gear.gear));
			if (itemSwap) {
				simUI.player.itemSwapSettings.setItemSwapSettings(
					eventID,
					true,
					simUI.sim.db.lookupItemSwap(itemSwap.itemSwap),
					Stats.fromProto(itemSwap.itemSwap.prepullBonusStats),
				);
			} else {
				simUI.player.itemSwapSettings.setEnableItemSwap(eventID, false);
			}
			if (talents) {
				simUI.player.setTalentsString(eventID, talents.data.talentsString);
			}
			if (rotationType && !rotation?.rotation.rotation) {
				simUI.player.aplRotation.type = rotationType;
				simUI.player.rotationChangeEmitter.emit(eventID);
			} else if (rotation?.rotation.rotation) {
				if (rotationType) simUI.player.aplRotation.type = rotationType;
				simUI.player.setAplRotation(eventID, rotation.rotation.rotation);
			}
			if (epWeights) simUI.player.setEpWeights(eventID, epWeights.epWeights);
			if (settings) {
				if (settings.race) simUI.player.setRace(eventID, settings.race);
				if (settings.partyBuffs) simUI.player.getParty()?.setBuffs(eventID, settings.partyBuffs);
				if (settings.consumables) simUI.player.setConsumes(eventID, settings.consumables);
				if (settings.playerOptions?.profession1) simUI.player.setProfession1(eventID, settings.playerOptions.profession1);
				if (settings.playerOptions?.profession2) simUI.player.setProfession2(eventID, settings.playerOptions.profession2);
				if (typeof settings.playerOptions?.distanceFromTarget === 'number')
					simUI.player.setDistanceFromTarget(eventID, settings.playerOptions.distanceFromTarget);
				if (settings.playerOptions?.reactionTimeMs) simUI.player.setReactionTime(eventID, settings.playerOptions.reactionTimeMs);
				if (settings.playerOptions?.channelClipDelayMs) simUI.player.setChannelClipDelay(eventID, settings.playerOptions.channelClipDelayMs);
				if (typeof settings.playerOptions?.inFrontOfTarget === 'boolean')
					simUI.player.setInFrontOfTarget(eventID, settings.playerOptions.inFrontOfTarget);
				if (settings.playerOptions?.enableItemSwap !== undefined && settings.playerOptions?.itemSwap) {
					simUI.player.itemSwapSettings.setItemSwapSettings(
						eventID,
						settings.playerOptions.enableItemSwap,
						simUI.sim.db.lookupItemSwap(settings.playerOptions.itemSwap),
						Stats.fromProto(settings.playerOptions.itemSwap.prepullBonusStats),
					);
				}
				if (settings.specOptions) {
					// Avoid object-spread over a large union type (produces an unassignable union);
					// getSpecOptions() already returns a copy, so mutating it is safe here.
					const mergedSpecOptions = simUI.player.getSpecOptions() as any;
					Object.assign(mergedSpecOptions, settings.specOptions);
					simUI.player.setSpecOptions(eventID, mergedSpecOptions);
				}
				if (settings.raidBuffs) simUI.sim.raid.setBuffs(eventID, settings.raidBuffs);
				if (settings.buffs) simUI.player.setBuffs(eventID, settings.buffs);
				if (settings.debuffs) simUI.sim.raid.setDebuffs(eventID, settings.debuffs);
				if (simUI.reforger && settings.reforgeSettings) {
					const { useCustomEpValues, statCaps, useSoftCapBreakpoints, freezeItemSlots, frozenItemSlots, breakpointLimits, maxGemPhase } =
						settings.reforgeSettings;

					if (useCustomEpValues) simUI.reforger.setUseCustomEPValues(eventID, useCustomEpValues);
					if (statCaps) simUI.reforger.setStatCaps(eventID, Stats.fromProto(statCaps));
					if (useSoftCapBreakpoints) simUI.reforger.setUseSoftCapBreakpoints(eventID, useSoftCapBreakpoints);
					if (freezeItemSlots) simUI.reforger.setFreezeItemSlots(eventID, freezeItemSlots);
					if (frozenItemSlots) simUI.reforger.setFrozenItemSlots(eventID, frozenItemSlots);
					if (breakpointLimits) simUI.reforger.setBreakpointLimits(eventID, Stats.fromProto(breakpointLimits));
					if (maxGemPhase) simUI.reforger.setMaxGemPhase(eventID, maxGemPhase);
				}
			}
			if (encounter) {
				if (encounter.encounter) simUI.sim.encounter.fromProto(eventID, encounter.encounter);
				if (encounter.healingModel) simUI.player.setHealingModel(eventID, encounter.healingModel);
				if (encounter.tanks) simUI.sim.raid.setTanks(eventID, encounter.tanks);
			}
		});
	}

	private isBuildActive({ gear, rotation, rotationType, talents, epWeights, encounter, settings }: PresetBuild): boolean {
		const types = this.types;
		const checkAll = !types;

		const hasGear = (checkAll || types.includes(PresetConfigurationCategory.Gear))
			? (gear ? EquipmentSpec.equals(gear.gear, this.simUI.player.getGear().asSpec()) : true)
			: true;
		const hasTalents = (checkAll || types.includes(PresetConfigurationCategory.Talents))
			? (talents
				? SavedTalents.equals(
						talents.data,
						SavedTalents.create({
							talentsString: this.simUI.player.getTalentsString(),
						}),
					)
				: true)
			: true;
		let hasRotation = true;
		if (checkAll || types.includes(PresetConfigurationCategory.Rotation) || types.includes(PresetConfigurationCategory.RotationType)) {
			if (rotationType) {
				hasRotation = rotationType === this.simUI.player.getRotationType();
			} else if (rotation?.rotation.rotation) {
				const activeRotation = this.simUI.player.getResolvedAplRotation();
				hasRotation = isEqualAPLRotation(this.simUI.player, activeRotation, rotation.rotation.rotation);
			}
		}
		const hasEpWeights = (checkAll || types.includes(PresetConfigurationCategory.EPWeights))
			? (epWeights ? this.simUI.player.getEpWeights().equals(epWeights.epWeights) : true)
			: true;

		let hasEncounter = true;
		let hasHealingModel = true;
		if (checkAll || types.includes(PresetConfigurationCategory.Encounter)) {
			hasEncounter = encounter?.encounter
				? Encounter.equals({ ...encounter.encounter, apiVersion: 0 }, { ...this.simUI.sim.encounter.toProto(), apiVersion: 0 })
				: true;
			hasHealingModel = encounter?.healingModel ? HealingModel.equals(encounter.healingModel, this.simUI.player.getHealingModel()) : true;
		}

		let hasRace = true;
		let hasProfession1 = true;
		let hasProfession2 = true;
		let hasDistanceFromTarget = true;
		let hasEnableItemSwap = true;
		let hasItemSwap = true;
		let hasSpecOptions = true;
		let hasConsumables = true;
		let hasPartyBuffs = true;
		let hasRaidBuffs = true;
		let hasBuffs = true;
		let hasDebuffs = true;
		if (checkAll || types.includes(PresetConfigurationCategory.Settings)) {
			hasRace = settings?.race ? this.simUI.player.getRace() === settings.race : true;
			hasProfession1 = settings?.playerOptions?.profession1 === undefined || this.simUI.player.getProfession1() === settings.playerOptions.profession1;
			hasProfession2 = settings?.playerOptions?.profession2 === undefined || this.simUI.player.getProfession2() === settings.playerOptions.profession2;
			hasDistanceFromTarget =
				settings?.playerOptions?.distanceFromTarget === undefined ||
				this.simUI.player.getDistanceFromTarget() === settings.playerOptions.distanceFromTarget;
			hasEnableItemSwap =
				settings?.playerOptions?.enableItemSwap === undefined ||
				this.simUI.player.itemSwapSettings.getEnableItemSwap() === settings.playerOptions.enableItemSwap;
			hasItemSwap =
				settings?.playerOptions?.itemSwap === undefined ||
				(!settings?.playerOptions?.enableItemSwap && !this.simUI.player.itemSwapSettings.getEnableItemSwap()) ||
				ItemSwap.equals(stripItemSwapApiVersion(this.simUI.player.itemSwapSettings?.toProto()), stripItemSwapApiVersion(settings?.playerOptions?.itemSwap));
			hasSpecOptions =
				settings?.specOptions && Object.keys(settings.specOptions).length
					? JSON.stringify(this.simUI.player.getSpecOptions()) == JSON.stringify(settings.specOptions)
					: true;
			hasConsumables = settings?.consumables ? ConsumesSpec.equals(this.simUI.player.getConsumes(), settings.consumables) : true;
			hasPartyBuffs = settings?.partyBuffs ? PartyBuffs.equals(this.simUI.player.getParty()?.getBuffs(), settings.partyBuffs) : true;
			hasRaidBuffs = settings?.raidBuffs ? RaidBuffs.equals(this.simUI.sim.raid.getBuffs(), settings.raidBuffs) : true;
			hasBuffs = settings?.buffs ? IndividualBuffs.equals(this.simUI.player.getBuffs(), settings.buffs) : true;
			hasDebuffs = settings?.debuffs ? Debuffs.equals(this.simUI.sim.raid.getDebuffs(), settings.debuffs) : true;
		}

		return (
			hasGear &&
			hasTalents &&
			hasRotation &&
			hasEpWeights &&
			hasEncounter &&
			hasHealingModel &&
			hasRace &&
			hasProfession1 &&
			hasProfession2 &&
			hasDistanceFromTarget &&
			hasEnableItemSwap &&
			hasItemSwap &&
			hasSpecOptions &&
			hasConsumables &&
			hasPartyBuffs &&
			hasRaidBuffs &&
			hasBuffs &&
			hasDebuffs
		);
	}
}

/** Strips apiVersion from an ItemSwap and its nested UnitStats so preset comparisons aren't version-sensitive. */
function stripItemSwapApiVersion(swap: ItemSwap | undefined): ItemSwap | undefined {
	if (!swap) return swap;
	return {
		...swap,
		prepullBonusStats: swap.prepullBonusStats ? { ...swap.prepullBonusStats, apiVersion: 0 } : swap.prepullBonusStats,
	};
}
