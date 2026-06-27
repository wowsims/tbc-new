import clsx from 'clsx';
import tippy, { hideAll } from 'tippy.js';
import { ref } from 'tsx-vanilla';

import i18n from '../../i18n/config.js';
import { SimSettingCategories } from '../constants/sim_settings';
import { IndividualSimUI } from '../individual_sim_ui';
import { Player } from '../player';
import { Player as PlayerProtoMessageType, ReforgeOptimizeMode, ReforgeOptimizeRequest, ReforgeSettings, StatCapType } from '../proto/api';
import { Class, Debuffs, GemColor, ItemQuality, ItemSlot, PartyBuffs, Profession, PseudoStat, RaidBuffs, Spec, Stat } from '../proto/common';
import { UIGem as Gem } from '../proto/ui';
import { ReforgeGearCache } from '../reforge_cache';
import { EquippedItem } from '../proto_utils/equipped_item';
import { Gear } from '../proto_utils/gear';
import { getEmptyGemSocketIconUrl } from '../proto_utils/gems';
import { statCapTypeNames } from '../proto_utils/names';
import { getGearKeyFromSpec } from '../proto_utils/utils';
import { translateItemQuality, translateSlotName } from '../../i18n/localization';
import { StatCap, Stats, UnitStat, UnitStatPresets } from '../proto_utils/stats';
import { ReforgeOptimizeConfig, Sim } from '../sim';
import { ActionGroupItem } from '../sim_ui';
import { RequestTypes } from '../sim_signal_manager';
import { EventID, TypedEvent } from '../typed_event';
import { distinct, isDevMode, phasesEnumToNumber } from '../utils';
import { BooleanPicker } from './pickers/boolean_picker';
import { EnumPicker } from './pickers/enum_picker';
import { NumberPicker, NumberPickerConfig } from './pickers/number_picker';
import { renderSavedEPWeights } from './saved_data_managers/ep_weights';
import Toast from './toast';
import { trackEvent, trackPageView } from '../../tracking/utils';
import { ProgressTrackerModal } from './progress_tracker_modal';
import { getEmptySlotIconUrl } from './gear_picker/utils';
import { CURRENT_PHASE, Phase } from '../constants/other';
import { CharacterStats } from './character_stats';

const INCLUDED_STATS: UnitStat[] = [
	UnitStat.fromStat(Stat.StatSpellHitRating),
	UnitStat.fromPseudoStat(PseudoStat.PseudoStatSchoolHitPercentArcane),
	UnitStat.fromPseudoStat(PseudoStat.PseudoStatSchoolHitPercentFire),
	UnitStat.fromPseudoStat(PseudoStat.PseudoStatSchoolHitPercentFrost),
	UnitStat.fromPseudoStat(PseudoStat.PseudoStatSchoolHitPercentHoly),
	UnitStat.fromPseudoStat(PseudoStat.PseudoStatSchoolHitPercentNature),
	UnitStat.fromPseudoStat(PseudoStat.PseudoStatSchoolHitPercentShadow),
	UnitStat.fromPseudoStat(PseudoStat.PseudoStatReducedCritTakenPercent),
	UnitStat.fromStat(Stat.StatSpellCritRating),
	UnitStat.fromStat(Stat.StatSpellHasteRating),
	UnitStat.fromStat(Stat.StatMeleeHitRating),
	UnitStat.fromPseudoStat(PseudoStat.PseudoStatMeleeHitPercent),
	UnitStat.fromStat(Stat.StatMeleeCritRating),
	UnitStat.fromStat(Stat.StatMeleeHasteRating),
	UnitStat.fromStat(Stat.StatExpertiseRating),
	UnitStat.fromStat(Stat.StatArmorPenetration),
	UnitStat.fromStat(Stat.StatDodgeRating),
	UnitStat.fromStat(Stat.StatParryRating),
	UnitStat.fromStat(Stat.StatDefenseRating),
	UnitStat.fromStat(Stat.StatResilienceRating),
];

type StatTooltipContent = { [key in Stat]?: () => Element | string };

const STAT_TOOLTIPS: StatTooltipContent = {
	[Stat.StatMeleeHasteRating]: () => (
		<>
			Final percentage value <strong>including</strong> all buffs/gear.
		</>
	),
	[Stat.StatSpellHasteRating]: () => (
		<>
			Final percentage value <strong>including</strong> all buffs/gear.
		</>
	),
};

export type ReforgeOptimizerOptions = {
	experimental?: true;
	statTooltips?: StatTooltipContent;
	statSelectionPresets?: UnitStatPresets[];
	// Allows you to enable breakpoint limits for Treshold type caps
	enableBreakpointLimits?: boolean;
	// Allows you to modify the stats before they are returned for the calculations
	// For example: Adding class specific Glyphs/Talents that are not added by the backend
	updateGearStatsModifier?: (baseStats: Stats) => Stats;
	// Allows you to get alternate default EPs
	// For example for Fury where you have SMF and TG EPs
	getEPDefaults?: (player: Player<any>) => Stats;
	// Allows you to modify default softCaps
	// For example you wish to add breakpoints for Berserking / Bloodlust if enabled
	updateSoftCaps?: (softCaps: StatCap[]) => StatCap[];
	// Allows you to specifiy additional information for the soft cap tooltips
	additionalSoftCapTooltipInformation?: StatTooltipContent;
	// Sets the default stat to be the highest for relative stat cap calculations
	// Defaults to Any
	defaultRelativeStatCap?: Stat | null;
};

export class ReforgeOptimizer {
	protected readonly simUI: IndividualSimUI<any>;
	protected readonly player: Player<any>;
	protected readonly playerClass: Class;
	protected readonly isExperimental: ReforgeOptimizerOptions['experimental'];
	protected readonly isHybridCaster: boolean;
	protected readonly isTankSpec: boolean;
	protected readonly sim: Sim;
	protected readonly defaults: IndividualSimUI<any>['individualConfig']['defaults'];
	protected reforgeDoneToast: Toast | null = null;
	protected getEPDefaults: ReforgeOptimizerOptions['getEPDefaults'];
	protected _statCaps: Stats = new Stats();
	protected breakpointLimits: Stats = new Stats();
	protected updateGearStatsModifier: ReforgeOptimizerOptions['updateGearStatsModifier'];
	protected _softCapsConfig: StatCap[];
	private useCustomEPValues = false;
	private useSoftCapBreakpoints = true;
	protected progressTrackerModal: ProgressTrackerModal;
	protected softCapBreakpoints: StatCap[] = [];
	protected updateSoftCaps: ReforgeOptimizerOptions['updateSoftCaps'];
	protected enableBreakpointLimits: ReforgeOptimizerOptions['enableBreakpointLimits'];
	protected statTooltips: StatTooltipContent = {};
	protected additionalSoftCapTooltipInformation: StatTooltipContent = {};
	protected statSelectionPresets: ReforgeOptimizerOptions['statSelectionPresets'];
	protected freezeItemSlots = false;
	protected frozenItemSlots = new Set<ItemSlot>();
	protected maxGemPhase = CURRENT_PHASE;
	protected maxGemQuality = ItemQuality.ItemQualityEpic;
	protected disableUniqueGems = false;
	protected undershootCaps = new Stats();
	protected isCancelling: boolean = false;
	protected previousGear: Gear | null = null;
	protected updatedGear: Gear | null = null;

	readonly statCapsChangeEmitter = new TypedEvent<void>('StatCaps');
	readonly useCustomEPValuesChangeEmitter = new TypedEvent<void>('UseCustomEPValues');
	readonly useSoftCapBreakpointsChangeEmitter = new TypedEvent<void>('UseSoftCapBreakpoints');
	readonly softCapBreakpointsChangeEmitter = new TypedEvent<void>('SoftCapBreakpoints');
	readonly breakpointLimitsChangeEmitter = new TypedEvent<void>('BreakpointLimits');
	readonly freezeItemSlotsChangeEmitter = new TypedEvent<void>('FreezeItemSlots');
	readonly maxGemPhaseEmitter = new TypedEvent<void>('MaxGemPhase');
	readonly maxGemQualityEmitter = new TypedEvent<void>('MaxGemQuality');
	readonly disableUniqueGemsChangeEmitter = new TypedEvent<void>('DisableUniqueGems');
	readonly undershootCapsChangeEmitter = new TypedEvent<void>('UndershootCaps');

	// Emits when any of the above emitters emit.
	readonly changeEmitter: TypedEvent<void>;

	constructor(simUI: IndividualSimUI<any>, options?: ReforgeOptimizerOptions) {
		this.simUI = simUI;
		this.player = simUI.player;
		this.playerClass = this.player.getClass();
		this.isExperimental = options?.experimental;
		this.isHybridCaster = [Spec.SpecBalanceDruid, Spec.SpecPriest, Spec.SpecElementalShaman].includes(this.player.getSpec());
		this.isTankSpec = this.player.getPlayerSpec().isTankSpec;
		this.sim = simUI.sim;
		this.defaults = simUI.individualConfig.defaults;
		this.getEPDefaults = options?.getEPDefaults;
		this.updateSoftCaps = options?.updateSoftCaps;
		this.updateGearStatsModifier = options?.updateGearStatsModifier;
		this._softCapsConfig = this.defaults.softCapBreakpoints || [];
		this.statTooltips = { ...STAT_TOOLTIPS, ...options?.statTooltips };
		this.additionalSoftCapTooltipInformation = { ...options?.additionalSoftCapTooltipInformation };
		this.statSelectionPresets = options?.statSelectionPresets;
		this._statCaps = this.defaults.statCaps || new Stats();
		this.enableBreakpointLimits = !!options?.enableBreakpointLimits;
		this.progressTrackerModal = new ProgressTrackerModal(simUI.rootElem, {
			id: 'reforge-optimizer-progress-tracker',
			title: 'Optimizing Gems',
			warning: (
				<>
					<p>
						Gemming can be a lengthy process, especially as specific stat caps and breakpoints come into play for classes. This may take a while,
						but be assured that the calculation will eventually complete.
					</p>
					<p className="mb-0">You may cancel this operation at any time using the button below.</p>
				</>
			),
			onCancel: async () => {
				this.isCancelling = true;
				if (isDevMode()) {
					console.log('User cancelled gem optimization');
				}
				try {
					await this.abortReforgeOptimization();
				} catch {}
				if (this.previousGear) this.player.setGear(TypedEvent.nextEventID(), this.previousGear);
				this.progressTrackerModal.hide();
				trackEvent({
					action: 'settings',
					category: 'reforging',
					label: 'suggest_cancel',
				});

				new Toast({
					variant: 'warning',
					body: i18n.t('sidebar.buttons.suggest_reforges.reforge_optimization_cancelled'),
					delay: 3000,
				});
			},
		});

		const startReforgeOptimizationEntry: ActionGroupItem = {
			label: i18n.t('sidebar.buttons.suggest_reforges.title'),
			cssClass: 'suggest-reforges-action-button flex-grow-1',
			onClick: async () => {
				this.reforgeDoneToast?.hide();
				this.reforgeDoneToast = null;

				this.progressTrackerModal.show();
				trackEvent({
					action: 'settings',
					category: 'reforging',
					label: 'suggest_start',
				});

				try {
					performance.mark('reforge-optimization-start');
					const gear = await this.optimizeReforges();
					await this.player.setGearAsync(TypedEvent.nextEventID(), gear);
					this.onReforgeDone();
				} catch (error) {
					if (this.isCancelling) return;
					this.onReforgeError(error);
				} finally {
					this.onReforgeFinally();
				}
			},
		};

		const contextMenuEntry: ActionGroupItem = {
			cssClass: 'suggest-reforges-button-settings',
			children: (
				<>
					<i className="fas fa-cog" />
				</>
			),
		};

		const {
			group,
			children: [startReforgeOptimizationButton, contextMenuButton],
		} = simUI.addActionGroup([startReforgeOptimizationEntry, contextMenuEntry], {
			cssClass: clsx('suggest-reforges-settings-group', this.isExperimental && !this.player.sim.getShowExperimental() && 'hide'),
		});

		this.bindToggleExperimental(group);

		tippy(startReforgeOptimizationButton, {
			theme: 'suggest-reforges-softcaps',
			placement: 'bottom',
			maxWidth: 310,
			interactive: true,
			onShow: instance => {
				if (!this.softCapsConfig?.length) return false;
				instance.setContent(this.buildReforgeButtonTooltip());
			},
		});

		tippy(contextMenuButton, {
			placement: 'bottom',
			content: i18n.t('sidebar.buttons.suggest_reforges.tooltip'),
		});

		this.buildContextMenu(contextMenuButton);

		this.changeEmitter = TypedEvent.onAny(
			[
				this.statCapsChangeEmitter,
				this.useCustomEPValuesChangeEmitter,
				this.useSoftCapBreakpointsChangeEmitter,
				this.softCapBreakpointsChangeEmitter,
				this.breakpointLimitsChangeEmitter,
				this.freezeItemSlotsChangeEmitter,
				this.maxGemPhaseEmitter,
				this.maxGemQualityEmitter,
				this.disableUniqueGemsChangeEmitter,
				this.undershootCapsChangeEmitter,
			],
			'ReforgeSettingsChange',
		);

		TypedEvent.onAny([this.useCustomEPValuesChangeEmitter, this.player.epWeightsChangeEmitter, this.statCapsChangeEmitter]).on(eventID => {
			if (this.useCustomEPValues && (this.player.hasCustomEPWeights() || !this._statCaps.equals(this.defaults.statCaps || new Stats()))) {
				this.setUseSoftCapBreakpoints(eventID, false);
			}
		});
	}

	private bindToggleExperimental(element: Element) {
		const toggle = () => element.classList[this.isExperimental && !this.player.sim.getShowExperimental() ? 'add' : 'remove']('hide');
		toggle();
		this.player.sim.showExperimentalChangeEmitter.on(() => {
			toggle();
		});
	}

	get softCapsConfig() {
		return this.updateSoftCaps?.(StatCap.cloneSoftCaps(this._softCapsConfig)) || this._softCapsConfig;
	}

	get softCapsConfigWithLimits() {
		if (!this.enableBreakpointLimits || !this.useSoftCapBreakpoints) return this.softCapsConfig;

		const softCaps = StatCap.cloneSoftCaps(this.softCapsConfig);
		for (const [unitStat, limit] of this.breakpointLimits.asUnitStatArray()) {
			if (!limit) continue;
			const config = softCaps.find(config => config.unitStat.equals(unitStat));
			const breakpointLimitExists = config?.breakpoints.some(breakpoint => breakpoint == limit);
			if (config && breakpointLimitExists) {
				config.breakpoints = config.breakpoints.filter(breakpoint => breakpoint <= limit);
				if (config.capType === StatCapType.TypeSoftCap) {
					config.postCapEPs = config.postCapEPs.slice(0, config.breakpoints.length);
				}
			}
		}
		return softCaps;
	}

	get preCapEPs(): Stats {
		let weights = this.player.getEpWeights();

		if (!this.useCustomEPValues) {
			if (this.getEPDefaults) {
				weights = this.getEPDefaults?.(this.player);
			} else if (this.player.hasCustomEPWeights()) {
				weights = this.defaults.epWeights;
			}
		}

		// Replace Spirit EP for hybrid casters with a small value in order to break ties between Spirit and Hit Reforges
		if (this.isHybridCaster) {
			weights = weights.withStat(Stat.StatSpirit, 0.01);
		}

		return weights;
	}

	buildReforgeButtonTooltip() {
		return (
			<>
				<p>{i18n.t('sidebar.buttons.suggest_reforges.breakpoints_implemented')}</p>
				<table className="w-100">
					<tbody>
						{this.softCapsConfigWithLimits?.map(({ unitStat, breakpoints, capType, postCapEPs }, index) => (
							<>
								<tr>
									<th className="text-nowrap" colSpan={2}>
										{unitStat.getShortName(this.playerClass)}
									</th>
									<td className="text-end">{statCapTypeNames.get(capType)}</td>
								</tr>
								{this.additionalSoftCapTooltipInformation[unitStat.getRootStat()] && (
									<>
										<tr>
											<td colSpan={3}>{this.additionalSoftCapTooltipInformation[unitStat.getRootStat()]?.()}</td>
										</tr>
										<tr>
											<td colSpan={3} className="pb-2"></td>
										</tr>
									</>
								)}
								<tr>
									<th className="text-end">
										<em>%</em>
									</th>
									<th colSpan={2} className="text-nowrap text-end">
										<em>{i18n.t('sidebar.buttons.suggest_reforges.post_cap_ep')}</em>
									</th>
								</tr>
								{breakpoints.map((breakpoint, breakpointIndex) => (
									<tr>
										<td className="text-end">{this.breakpointValueToDisplayPercentage(breakpoint, unitStat)}</td>
										<td colSpan={2} className="text-end">
											{unitStat
												.convertEpToRatingScale(capType === StatCapType.TypeThreshold ? postCapEPs[0] : postCapEPs[breakpointIndex])
												.toFixed(2)}
										</td>
									</tr>
								))}
								{index !== this.softCapsConfigWithLimits.length - 1 && (
									<>
										<tr>
											<td colSpan={3} className="border-bottom pb-2"></td>
										</tr>
										<tr>
											<td colSpan={3} className="pb-2"></td>
										</tr>
									</>
								)}
							</>
						))}
					</tbody>
				</table>
			</>
		);
	}

	setStatCaps(eventID: EventID, newStatCaps: Stats) {
		this._statCaps = newStatCaps;
		this.statCapsChangeEmitter.emit(eventID);
	}

	get statCaps() {
		return this.useCustomEPValues ? this._statCaps : this.defaults.statCaps || new Stats();
	}

	setUseCustomEPValues(eventID: EventID, newUseCustomEPValues: boolean) {
		if (newUseCustomEPValues !== this.useCustomEPValues) {
			this.useCustomEPValues = newUseCustomEPValues;
			this.useCustomEPValuesChangeEmitter.emit(eventID);
		}
	}

	setUseSoftCapBreakpoints(eventID: EventID, newUseSoftCapBreakpoints: boolean) {
		if (newUseSoftCapBreakpoints !== this.useSoftCapBreakpoints) {
			this.useSoftCapBreakpoints = newUseSoftCapBreakpoints;
			this.useSoftCapBreakpointsChangeEmitter.emit(eventID);
		}
	}

	setBreakpointLimits(eventID: EventID, newLimits: Stats) {
		this.breakpointLimits = newLimits;
		this.breakpointLimitsChangeEmitter.emit(eventID);
	}

	setSoftCapBreakpoints(eventID: EventID, newSoftCapBreakpoints: StatCap[]) {
		this.softCapBreakpoints = newSoftCapBreakpoints;
		this.softCapBreakpointsChangeEmitter.emit(eventID);
	}

	setFreezeItemSlots(eventID: EventID, newValue: boolean) {
		if (this.freezeItemSlots !== newValue) {
			this.freezeItemSlots = newValue;
			this.frozenItemSlots.clear();
			this.freezeItemSlotsChangeEmitter.emit(eventID);
		}
	}

	setFrozenItemSlot(eventID: EventID, slot: ItemSlot, frozen: boolean) {
		if (this.getFrozenItemSlot(slot) !== frozen) {
			this.frozenItemSlots[frozen ? 'add' : 'delete'](slot);
			this.freezeItemSlotsChangeEmitter.emit(eventID);
		}
	}

	// Sets all frozen item slots at once
	setFrozenItemSlots(eventID: EventID, slots: ItemSlot[]) {
		this.frozenItemSlots.clear();
		slots.forEach(slot => this.frozenItemSlots.add(slot));
		this.freezeItemSlotsChangeEmitter.emit(eventID);
	}

	getFrozenItemSlot(slot: ItemSlot): boolean {
		return this.frozenItemSlots.has(slot);
	}

	setMaxGemPhase(eventID: EventID, phase: number): void {
		this.maxGemPhase = phase;
		this.maxGemPhaseEmitter.emit(eventID);
	}

	getMaxGemPhase(): number {
		return this.maxGemPhase;
	}

	setMaxGemQuality(eventID: EventID, quality: ItemQuality): void {
		this.maxGemQuality = quality;
		this.maxGemQualityEmitter.emit(eventID);
	}

	getMaxGemQuality(): ItemQuality {
		return this.maxGemQuality;
	}

	setDisableUniqueGems(eventID: EventID, disableUniqueGems: boolean): void {
		this.disableUniqueGems = disableUniqueGems;
		this.disableUniqueGemsChangeEmitter.emit(eventID);
	}

	buildContextMenu(button: HTMLButtonElement) {
		const instance = tippy(button, {
			interactive: true,
			trigger: 'click',
			theme: 'reforge-optimiser-popover',
			placement: 'right-start',
			onShow: instance => {
				trackPageView('Reforge Settings', 'reforge-settings');

				const useCustomEPValuesInput = new BooleanPicker(null, this.player, {
					extraCssClasses: ['mb-2'],
					id: 'reforge-optimizer-enable-custom-ep-weights',
					label: i18n.t('sidebar.buttons.suggest_reforges.use_custom'),
					inline: true,
					changedEvent: () => this.useCustomEPValuesChangeEmitter,
					getValue: () => this.useCustomEPValues,
					setValue: (eventID, _player, newValue) => {
						trackEvent({
							action: 'settings',
							category: 'reforging',
							label: 'use_custom_ep',
							value: newValue,
						});
						this.setUseCustomEPValues(eventID, newValue);
					},
				});

				let useSoftCapBreakpointsInput: BooleanPicker<Player<any>> | null = null;
				if (this.softCapsConfig?.length) {
					useSoftCapBreakpointsInput = new BooleanPicker(null, this.player, {
						extraCssClasses: ['mb-2'],
						id: 'reforge-optimizer-enable-soft-cap-breakpoints',
						label: i18n.t('sidebar.buttons.suggest_reforges.use_soft_cap_breakpoints'),
						inline: true,
						changedEvent: () => this.useSoftCapBreakpointsChangeEmitter,
						getValue: () => this.useSoftCapBreakpoints,
						setValue: (eventID, _player, newValue) => {
							trackEvent({
								action: 'settings',
								category: 'reforging',
								label: 'softcap_breakpoints',
								value: newValue,
							});
							this.setUseSoftCapBreakpoints(eventID, newValue);
						},
					});
				}

				const disableUniqueGems = new BooleanPicker(null, this.player, {
					extraCssClasses: ['mb-2'],
					id: 'reforge-optimizer-disable-unique-gems',
					label: i18n.t('sidebar.buttons.suggest_reforges.disable_unique_gems'),
					inline: true,
					changedEvent: () => this.disableUniqueGemsChangeEmitter,
					getValue: () => this.disableUniqueGems,
					setValue: (eventID, _player, newValue) => {
						trackEvent({
							action: 'settings',
							category: 'reforging',
							label: 'disable_unique_gems',
							value: newValue,
						});
						this.setDisableUniqueGems(eventID, newValue);
					},
				});

				const maxGemPhaseInput = new EnumPicker(null, this.player, {
					extraCssClasses: ['mb-2'],
					id: 'reforge-optimizer-max-gem-phase',
					label: i18n.t('sidebar.buttons.suggest_reforges.max_gem_phase'),
					defaultValue: this.maxGemPhase,
					values: phasesEnumToNumber().map(phaseIndex => ({
						name: i18n.t(`common.phases.${phaseIndex}`),
						value: phaseIndex,
					})),
					changedEvent: () => this.maxGemPhaseEmitter,
					getValue: () => this.maxGemPhase,
					setValue: (_eventID, _player, newValue) => {
						trackEvent({
							action: 'settings',
							category: 'reforging',
							label: 'max_gem_phase',
							value: newValue,
						});
						this.setMaxGemPhase(TypedEvent.nextEventID(), newValue);
					},
				});

				const maxGemQualityInput = new EnumPicker(null, this.player, {
					extraCssClasses: ['mb-2'],
					id: 'reforge-optimizer-max-gem-quality',
					label: i18n.t('sidebar.buttons.suggest_reforges.max_gem_quality'),
					defaultValue: this.maxGemQuality,
					values: Object.values(ItemQuality)
						.filter((q): q is number => typeof q === 'number' && q >= ItemQuality.ItemQualityUncommon && q <= ItemQuality.ItemQualityEpic)
						.map(quality => ({
							name: translateItemQuality(quality),
							value: quality,
						})),
					changedEvent: () => this.maxGemQualityEmitter,
					getValue: () => this.maxGemQuality,
					setValue: (_eventID, _player, newValue) => {
						trackEvent({
							action: 'settings',
							category: 'reforging',
							label: 'max_gem_quality',
							value: newValue,
						});
						this.setMaxGemQuality(TypedEvent.nextEventID(), newValue);
					},
				});

				const freezeItemSlotsInput = new BooleanPicker(null, this.player, {
					extraCssClasses: ['mb-2'],
					id: 'reforge-optimizer-freeze-item-slots',
					label: i18n.t('sidebar.buttons.suggest_reforges.freeze_item_slots'),
					labelTooltip: i18n.t('sidebar.buttons.suggest_reforges.freeze_item_slots_tooltip'),
					inline: true,
					changedEvent: () => this.freezeItemSlotsChangeEmitter,
					getValue: () => this.freezeItemSlots,
					setValue: (eventID, _player, newValue) => {
						trackEvent({
							action: 'settings',
							category: 'reforging',
							label: 'freeze_item_slots',
							value: newValue,
						});
						this.setFreezeItemSlots(eventID, newValue);
					},
				});

				const descriptionRef = ref<HTMLParagraphElement>();
				instance.setContent(
					<>
						{useCustomEPValuesInput.rootElem}
						<div ref={descriptionRef} className={clsx('mb-0', this.useCustomEPValues && 'hide')}>
							<p>{i18n.t('sidebar.buttons.suggest_reforges.enable_modification')}</p>
							<p>{i18n.t('sidebar.buttons.suggest_reforges.modify_in_editor')}</p>
							<p>{i18n.t('sidebar.buttons.suggest_reforges.hard_cap_info')}</p>
						</div>
						{this.buildCapsList({
							useCustomEPValuesInput: useCustomEPValuesInput,
							description: descriptionRef.value!,
						})}
						{useSoftCapBreakpointsInput?.rootElem}
						{this.buildSoftCapBreakpointsLimiter({ useSoftCapBreakpointsInput })}
						{disableUniqueGems.rootElem}
						{maxGemPhaseInput.rootElem}
						{maxGemQualityInput.rootElem}
						{freezeItemSlotsInput.rootElem}
						{this.buildFrozenSlotsInputs()}
						{this.buildEPWeightsToggle()}
					</>,
				);
			},
			onHidden: () => {
				instance.setContent(<></>);
			},
		});
	}

	buildFrozenSlotsInputs() {
		const allSlots = this.player.getGear().getItemSlots();
		const numRows = Math.floor(allSlots.length / 2) + 1;
		const slotsByRow: ItemSlot[][] = [];

		for (let rowIdx = 0; rowIdx < numRows; rowIdx++) {
			slotsByRow.push(allSlots.slice(rowIdx * 2, (rowIdx + 1) * 2));
		}

		const tableRef = ref<HTMLTableElement>();
		const content = (
			<table className={clsx('mb-2', { 'd-none': !this.freezeItemSlots })} ref={tableRef}>
				{slotsByRow.map(slots => {
					const rowRef = ref<HTMLTableRowElement>();
					const row = (
						<tr ref={rowRef}>
							{slots.map(slot => {
								const picker = new BooleanPicker(null, this.player, {
									id: 'reforge-optimizer-freeze-' + ItemSlot[slot],
									label: translateSlotName(slot),
									inline: true,
									changedEvent: () => this.freezeItemSlotsChangeEmitter,
									getValue: () => this.getFrozenItemSlot(slot) || false,
									setValue: (eventID, _player, newValue) => {
										this.setFrozenItemSlot(eventID, slot, newValue);
									},
								});
								const column = <td>{picker.rootElem}</td>;
								return column;
							})}
						</tr>
					);
					return row;
				})}
			</table>
		);

		this.freezeItemSlotsChangeEmitter.on(() => {
			tableRef.value?.classList[this.freezeItemSlots ? 'remove' : 'add']('d-none');
		});

		return content;
	}

	buildCapsList({ useCustomEPValuesInput, description }: { useCustomEPValuesInput: BooleanPicker<Player<any>>; description: HTMLElement }) {
		const sharedInputConfig: Pick<NumberPickerConfig<Player<any>>, 'changedEvent'> = {
			changedEvent: _ => TypedEvent.onAny([this.useSoftCapBreakpointsChangeEmitter, this.statCapsChangeEmitter]),
		};

		const tableRef = ref<HTMLTableElement>();
		const statCapTooltipRef = ref<HTMLButtonElement>();
		const defaultStatCapsButtonRef = ref<HTMLButtonElement>();

		const content = (
			<table ref={tableRef} className={clsx('reforge-optimizer-stat-cap-table mb-2', !this.useCustomEPValues && 'hide')}>
				<thead>
					<tr>
						<th colSpan={4} className="pb-3">
							<div className="d-flex">
								<h6 className="content-block-title mb-0 me-1">{i18n.t('sidebar.buttons.suggest_reforges.edit_stat_caps')}</h6>
								<button ref={statCapTooltipRef} className="d-inline">
									<i className="fa-regular fa-circle-question" />
								</button>
								<button
									ref={defaultStatCapsButtonRef}
									className="d-inline ms-auto"
									onclick={() => this.setStatCaps(TypedEvent.nextEventID(), this.defaults.statCaps || new Stats())}>
									<i className="fas fa-arrow-rotate-left" />
								</button>
							</div>
						</th>
					</tr>
					<tr>
						<th>{i18n.t('sidebar.buttons.suggest_reforges.stat')}</th>
						<th colSpan={3} className="text-end">
							%
						</th>
						<th colSpan={1} className="text-start">
							Max?
						</th>
					</tr>
				</thead>
				<tbody>
					{this.simUI.individualConfig.displayStats.map(unitStat => {
						const rootStat = unitStat.hasRootStat() ? unitStat.getRootStat() : null;
						if (!INCLUDED_STATS.some(us => us.equals(unitStat))) return;

						const listElementRef = ref<HTMLTableRowElement>();
						const statName = unitStat.getShortName(this.player.getClass());

						const sharedStatInputConfig: Pick<NumberPickerConfig<Player<any>>, 'getValue' | 'setValue'> = {
							getValue: () => {
								return this.toVisualUnitStatPercentage(this.statCaps.getUnitStat(unitStat), unitStat);
							},
							setValue: (_eventID, _player, newValue) => {
								this.setStatCaps(
									TypedEvent.nextEventID(),
									this.statCaps.withUnitStat(unitStat, this.toDefaultUnitStatValue(newValue, unitStat)),
								);
							},
						};

						const percentagePicker = new NumberPicker(null, this.player, {
							id: `reforge-optimizer-${statName}-percentage`,
							float: true,
							maxDecimalDigits: 5,
							showZeroes: false,
							positive: true,
							extraCssClasses: ['mb-0'],
							enableWhen: () => this.isAllowedToOverrideStatCaps || !this.softCapsConfig.some(config => config.unitStat.equals(unitStat)),
							...sharedInputConfig,
							...sharedStatInputConfig,
						});

						const undershootPicker = new BooleanPicker(null, this.player, {
							id: `reforge-optimizer-${statName}-undershoot`,
							label: '',
							inline: false,
							changedEvent: () => this.undershootCapsChangeEmitter,
							getValue: () => this.undershootCaps.getUnitStat(unitStat) > 0,
							setValue: (_eventID, _player, newValue) => {
								this.undershootCaps = this.undershootCaps.withUnitStat(unitStat, newValue ? 1 : 0);
							},
						});

						const statPresets = this.statSelectionPresets?.find(entry => entry.unitStat.equals(unitStat))?.presets;

						const presets = !!statPresets
							? new EnumPicker(null, this.player, {
									id: `reforge-optimizer-${statName}-presets`,
									extraCssClasses: ['mb-0'],
									label: '',
									values: [
										{ name: i18n.t('sidebar.buttons.suggest_reforges.select_preset'), value: 0 },
										...[...statPresets.keys()].map(key => {
											const percentValue = statPresets.get(key)!;

											return {
												name: `${key} - ${percentValue.toFixed(2)}%`,
												value: percentValue,
											};
										}),
									].sort((a, b) => a.value - b.value),
									enableWhen: () => this.isAllowedToOverrideStatCaps || !this.softCapsConfig.some(config => config.unitStat.equals(unitStat)),
									...sharedInputConfig,
									...sharedStatInputConfig,
								})
							: null;

						const tooltipText = rootStat !== null ? this.statTooltips[rootStat] : null;
						const statTooltipRef = ref<HTMLButtonElement>();

						const row = (
							<>
								<tr ref={listElementRef} className="reforge-optimizer-stat-cap-item">
									<td>
										<div className="reforge-optimizer-stat-cap-item-label">
											{statName}{' '}
											{tooltipText && (
												<button ref={statTooltipRef} className="d-inline">
													<i className="fa-regular fa-circle-question" />
												</button>
											)}
										</div>
									</td>
									<td colSpan={3}>{percentagePicker.rootElem}</td>
									<td colSpan={1} className="text-end">
										{undershootPicker.rootElem}
									</td>
								</tr>
								{presets && (
									<tr>
										<td></td>
										<td colSpan={3}>{presets.rootElem}</td>
									</tr>
								)}
							</>
						);

						const tooltip = tooltipText
							? tippy(statTooltipRef.value!, {
									content: tooltipText,
								})
							: null;

						useCustomEPValuesInput.addOnDisposeCallback(() => tooltip?.destroy());

						return row;
					})}
				</tbody>
			</table>
		);

		if (statCapTooltipRef.value) {
			const tooltip = tippy(statCapTooltipRef.value, {
				content: i18n.t('sidebar.buttons.suggest_reforges.stat_caps_tooltip'),
			});
			useCustomEPValuesInput.addOnDisposeCallback(() => tooltip.destroy());
		}
		if (defaultStatCapsButtonRef.value) {
			const tooltip = tippy(defaultStatCapsButtonRef.value, {
				content: i18n.t('sidebar.buttons.suggest_reforges.reset_to_defaults'),
			});
			useCustomEPValuesInput.addOnDisposeCallback(() => tooltip.destroy());
		}

		const event = this.useCustomEPValuesChangeEmitter.on(() => {
			tableRef.value?.classList[this.useCustomEPValues ? 'remove' : 'add']('hide');
			description?.classList[!this.useCustomEPValues ? 'remove' : 'add']('hide');
		});

		useCustomEPValuesInput.addOnDisposeCallback(() => {
			content.remove();
			event.dispose();
			this.undershootCaps = new Stats();
		});

		return content;
	}

	buildEPWeightsToggle() {
		const epWeightsContainerRef = ref<HTMLDivElement>();
		const content = (
			<>
				<div ref={epWeightsContainerRef} />
				{this.simUI.epWeightsModal && (
					<button
						className="btn btn-outline-primary mt-2"
						onclick={() => {
							this.simUI.epWeightsModal?.open();
							hideAll();
						}}>
						{i18n.t('sidebar.buttons.suggest_reforges.edit_weights')}
					</button>
				)}
			</>
		);

		const render = () => {
			const container = epWeightsContainerRef.value;
			if (container) {
				const epPicker = renderSavedEPWeights(null, this.simUI, {
					extraCssClasses: ['mt-3'],
					loadOnly: true,
					presetsOnly: !this.useCustomEPValues,
				});
				container.replaceChildren(epPicker.rootElem);
			}
		};

		this.useCustomEPValuesChangeEmitter.on(() => render());
		render();

		return content;
	}

	buildSoftCapBreakpointsLimiter({ useSoftCapBreakpointsInput }: { useSoftCapBreakpointsInput: BooleanPicker<Player<any>> | null }) {
		if (!this.enableBreakpointLimits || !useSoftCapBreakpointsInput) return null;

		const tableRef = ref<HTMLTableElement>();
		const breakpointsLimitTooltipRef = ref<HTMLButtonElement>();

		const content = (
			<table ref={tableRef} className={clsx('reforge-optimizer-stat-cap-table mb-2', !this.useSoftCapBreakpoints && 'hide')}>
				<thead>
					<tr>
						<th colSpan={3} className="pb-3">
							<div className="d-flex">
								<h6 className="content-block-title mb-0 me-1">{i18n.t('sidebar.buttons.suggest_reforges.breakpoint_limit')}</h6>
								<button ref={breakpointsLimitTooltipRef} className="d-inline">
									<i className="fa-regular fa-circle-question" />
								</button>
							</div>
						</th>
					</tr>
				</thead>
				<tbody>
					{this.softCapsConfig
						.filter(
							config =>
								(config.capType === StatCapType.TypeThreshold || config.capType === StatCapType.TypeSoftCap) && config.breakpoints.length > 0,
						)
						.map(({ breakpoints, unitStat }) => {
							if (!INCLUDED_STATS.some(us => us.equals(unitStat))) return;

							const listElementRef = ref<HTMLTableRowElement>();
							const statName = unitStat.getShortName(this.player.getClass());
							const picker = breakpoints
								? new EnumPicker(null, this.player, {
										id: `reforge-optimizer-${statName}-presets`,
										extraCssClasses: ['mb-0'],
										label: '',
										values: [
											{ name: i18n.t('sidebar.buttons.suggest_reforges.no_limit_set'), value: 0 },
											...breakpoints.map(breakpoint => ({
												name: `${this.breakpointValueToDisplayPercentage(breakpoint, unitStat)}%`,
												value: breakpoint,
											})),
										].sort((a, b) => a.value - b.value),
										changedEvent: _ => TypedEvent.onAny([this.useSoftCapBreakpointsChangeEmitter]),
										getValue: () => {
											const breakpointLimits = this.breakpointLimits;
											let limit = breakpointLimits.getUnitStat(unitStat);
											if (!breakpoints.some(breakpoint => breakpoint == limit)) {
												limit = 0;
											}

											return limit;
										},
										setValue: (eventID, _player, newValue) => {
											this.setBreakpointLimits(eventID, this.breakpointLimits.withUnitStat(unitStat, newValue));
										},
									})
								: null;

							if (!picker?.rootElem) return null;

							const row = (
								<>
									<tr ref={listElementRef} className="reforge-optimizer-stat-cap-item">
										<td>
											<div className="reforge-optimizer-stat-cap-item-label">{statName}</div>
										</td>
										<td colSpan={2}>{picker.rootElem}</td>
									</tr>
								</>
							);

							return row;
						})}
				</tbody>
			</table>
		);

		if (breakpointsLimitTooltipRef.value) {
			const tooltip = tippy(breakpointsLimitTooltipRef.value, {
				content: i18n.t('sidebar.buttons.suggest_reforges.breakpoint_limit_tooltip'),
			});
			useSoftCapBreakpointsInput.addOnDisposeCallback(() => tooltip.destroy());
		}

		const event = this.useSoftCapBreakpointsChangeEmitter.on(() => {
			const isUsingBreakpoints = this.useSoftCapBreakpoints;
			tableRef.value?.classList[isUsingBreakpoints ? 'remove' : 'add']('hide');
		});

		useSoftCapBreakpointsInput.addOnDisposeCallback(() => {
			content.remove();
			event?.dispose();
		});

		return content;
	}

	get isAllowedToOverrideStatCaps() {
		return !(this.useSoftCapBreakpoints && this.softCapsConfig);
	}

	get processedStatCaps() {
		let statCaps = this.statCaps;
		if (!this.isAllowedToOverrideStatCaps)
			this.softCapsConfigWithLimits.forEach(({ unitStat }) => {
				statCaps = statCaps.withUnitStat(unitStat, 0);
			});

		return statCaps;
	}

	getReforgeOptimizeConfig(gear: Gear): ReforgeOptimizeConfig {
		const settings = this.toProto();
		settings.statCaps = this.processedStatCaps.toProto();

		return {
			gear,
			preCapEPWeights: this.preCapEPs,
			undershootCaps: this.undershootCaps,
			settings,
			softCaps: this.softCapsConfigWithLimits,
		};
	}

	static async getConfigHash({
		player,
		reforgeRequest,
		raidBuffs,
		partyBuffs,
		debuffs,
	}: {
		player: Player<any>;
		reforgeRequest: ReforgeOptimizeRequest;
		raidBuffs: RaidBuffs;
		partyBuffs: PartyBuffs | undefined;
		debuffs: Debuffs;
	}): Promise<string> {
		const playerProto = player.toProto(true, false, [
			SimSettingCategories.Talents,
			SimSettingCategories.Consumes,
			SimSettingCategories.External,
			SimSettingCategories.Miscellaneous,
		]);
		playerProto.equipment = undefined;
		playerProto.database = undefined;
		playerProto.channelClipDelayMs = 0;
		playerProto.inFrontOfTarget = false;
		playerProto.distanceFromTarget = 0;
		playerProto.healingModel = undefined;

		const reforgeOptimizerConfigForHash = ReforgeOptimizeRequest.clone(reforgeRequest);
		reforgeOptimizerConfigForHash.requestId = '';
		reforgeOptimizerConfigForHash.raid = undefined;
		reforgeOptimizerConfigForHash.debug = false;
		reforgeOptimizerConfigForHash.mode = ReforgeOptimizeMode.ReforgeOptimizeModeSingle;
		reforgeOptimizerConfigForHash.gemOptions = reforgeOptimizerConfigForHash.gemOptions.sort((a, b) => a.id - b.id);

		return ReforgeGearCache.getHash({
			player: PlayerProtoMessageType.toJsonString(playerProto),
			raid: {
				buffs: RaidBuffs.toJsonString(raidBuffs),
				partyBuffs: partyBuffs ? PartyBuffs.toJsonString(partyBuffs) : null,
				debuffs: Debuffs.toJsonString(debuffs),
			},
			optimizer: ReforgeOptimizeRequest.toJsonString(reforgeOptimizerConfigForHash),
		});
	}

	getReforgeRequestForHash(config: ReforgeOptimizeConfig): ReforgeOptimizeRequest {
		return ReforgeOptimizeRequest.create({
			...ReforgeOptimizer.makeReforgeConfigRequestFields(config, this.sim.db),
		});
	}

	async optimizeReforges(gear?: Gear) {
		if (isDevMode()) console.log('Starting backend reforge optimization...');

		const previousGear = gear || this.player.getGear();
		this.previousGear = previousGear;
		const config = this.getReforgeOptimizeConfig(previousGear);
		const cache = ReforgeGearCache.get(this.player.getPlayerSpec());
		const configHash = await ReforgeOptimizer.getConfigHash({
			player: this.player,
			reforgeRequest: this.getReforgeRequestForHash(config),
			raidBuffs: this.sim.raid.getBuffs(),
			partyBuffs: this.player.getParty()?.getBuffs(),
			debuffs: this.sim.raid.getDebuffs(),
		});
		const frozenItemSlots = config.settings.freezeItemSlots && config.settings.frozenItemSlots.length ? config.settings.frozenItemSlots : undefined;
		const cacheKey = await ReforgeGearCache.getKey(getGearKeyFromSpec(previousGear.asSpec(), frozenItemSlots), configHash);
		const cachedGear = await cache.get(cacheKey);
		if (cachedGear) {
			if (isDevMode()) console.log('Reforge optimization: cache hit.');
			return this.sim.db.lookupEquipmentSpec(cachedGear);
		}

		const result = await this.sim.reforgeOptimize(config);

		if (result.error) {
			throw new Error(result.error.message || 'Backend reforge optimization failed.');
		}
		if (!result.optimizedGear) {
			throw new Error('Backend reforge optimizer did not return optimized gear.');
		}

		await cache.setGear(cacheKey, result.optimizedGear);

		const optimizedGear = this.sim.db.lookupEquipmentSpec(result.optimizedGear);
		this.updatedGear = optimizedGear;

		return optimizedGear;
	}

	static getReforgeGemOptions(db: typeof Sim.prototype.db): Gem[] {
		return distinct(
			[GemColor.GemColorRed, GemColor.GemColorBlue, GemColor.GemColorYellow]
				.flatMap(socketColor => db.getGems(socketColor))
				.filter(gem => !gem.name.includes('Perfect') && gem.quality >= ItemQuality.ItemQualityRare)
				.flat(),
			(a, b) => a.id == b.id,
		);
	}

	static makeReforgeConfigRequestFields(config: ReforgeOptimizeConfig, db: typeof Sim.prototype.db) {
		return {
			preCapEpWeights: config.preCapEPWeights.toProto(),
			undershootCaps: config.undershootCaps.toProto(),
			settings: config.settings,
			softCaps: config.softCaps.map(softCap => ({
				unitStat: softCap.unitStat.toProto(),
				breakpoints: softCap.breakpoints.slice(),
				capType: softCap.capType,
				postCapEPs: softCap.postCapEPs.slice(),
			})),
			gemOptions: ReforgeOptimizer.getReforgeGemOptions(db).map(gem => ({
				id: gem.id,
				name: gem.name,
				icon: gem.icon,
				color: gem.color,
				stats: gem.stats.slice(),
				phase: gem.phase,
				quality: gem.quality ?? ItemQuality.ItemQualityJunk,
				unique: gem.unique,
				requiredProfession: gem.requiredProfession ?? Profession.ProfessionUnknown,
			})),
		};
	}

	async updateGear(gear: Gear): Promise<Stats> {
		const currentStats = await this.sim.getCharacterStatsForGear(TypedEvent.nextEventID(), gear);
		let baseStats = Stats.fromProto(currentStats.finalStats);
		baseStats = baseStats.add(CharacterStats.getDebuffStats(this.player));
		if (this.updateGearStatsModifier) baseStats = this.updateGearStatsModifier(baseStats);
		return baseStats;
	}

	computeReforgeSoftCaps(baseStats: Stats): StatCap[] {
		const reforgeSoftCaps: StatCap[] = [];

		if (!this.isAllowedToOverrideStatCaps) {
			this.softCapsConfigWithLimits.slice().forEach(config => {
				let weights = config.postCapEPs.slice();
				const relativeBreakpoints = [];

				for (const breakpoint of config.breakpoints) {
					relativeBreakpoints.push(baseStats.computeGapToCap(config.unitStat, breakpoint));
				}

				// For stats that are configured as thresholds rather than soft caps,
				// reverse the order of evaluation of the breakpoints so that the
				// largest relevant threshold is always targeted. Likewise, use a
				// single value for the post-cap EP for these stats, which should be
				// interpreted (and computed) as the residual stat value just after
				// passing a threshold discontinuity.
				if (config.capType == StatCapType.TypeThreshold) {
					relativeBreakpoints.reverse();
					weights = Array(relativeBreakpoints.length).fill(weights[0]);
				}

				reforgeSoftCaps.push(new StatCap(config.unitStat, relativeBreakpoints, config.capType, weights));
			});
		}

		return reforgeSoftCaps;
	}

	private toVisualUnitStatPercentage(statValue: number, unitStat: UnitStat) {
		return unitStat.convertDefaultUnitsToPercent(statValue)!;
	}

	private toDefaultUnitStatValue(value: number, unitStat: UnitStat) {
		return unitStat.convertPercentToDefaultUnits(value)!;
	}

	private breakpointValueToDisplayPercentage(value: number, unitStat: UnitStat) {
		return unitStat.convertDefaultUnitsToPercent(value)!.toFixed(2);
	}

	onReforgeDone() {
		const currentGear = this.player.getGear();
		const itemSlots = currentGear.getItemSlots();
		const changedSlots = new Map<ItemSlot, EquippedItem | undefined>();
		for (const slot of itemSlots) {
			const prev = this.previousGear?.getEquippedItem(slot);
			const current = currentGear?.getEquippedItem(slot);

			if ((!prev && current) || (prev && current && !prev?.equals(current))) changedSlots.set(slot, current);
		}
		const hasReforgeChanges = changedSlots.size;

		const changedReforgeMessage = (
			<>
				<p className="mb-0">{i18n.t('gear_tab.reforge_success.title')}</p>
				<ul className="suggest-reforges-gear-list list-reset">
					{itemSlots.map(slot => {
						const item = changedSlots.get(slot);
						const slotName = translateSlotName(slot);
						const iconRef = ref<HTMLDivElement>();
						const reforgeRef = ref<HTMLDivElement>();
						const socketsContainerRef = ref<HTMLDivElement>();
						const itemElement = (
							<div className="item-picker-root">
								<div
									ref={iconRef}
									className="item-picker-icon-wrapper"
									style={{
										backgroundImage: `url('${getEmptySlotIconUrl(slot)}')`,
									}}>
									<div ref={reforgeRef} className="suggest-reforges-gear-reforge interactive d-none"></div>
									<div ref={socketsContainerRef} className="item-picker-sockets-container"></div>
								</div>
							</div>
						);

						if (item) {
							item.asActionId()
								.fill(undefined)
								.then(filledId => {
									filledId.setBackground(iconRef.value!);
								});

							const previousItem = this.previousGear?.getEquippedItem(slot);
							const previousGems = previousItem?.gems;

							const { gems } = item;

							if (gems || previousGems) {
								const changedGems: number[] = [];
								previousItem?.gemSockets.forEach((_, socketIdx) => {
									const previousGem = previousGems ? previousGems[socketIdx] : undefined;
									const currentGem = gems ? gems[socketIdx] : undefined;
									if (previousGem?.id !== currentGem?.id) {
										changedGems.push(socketIdx);
									}
								});

								item.allSocketColors().forEach((socketColor, gemIdx) => {
									const hasChangedSocket = changedGems.includes(gemIdx);
									const socketRef = ref<HTMLDivElement>();
									const gemName = gems[gemIdx]?.name;
									socketsContainerRef.value?.appendChild(
										<div
											ref={socketRef}
											className={clsx('gem-socket-container', hasChangedSocket && 'interactive')}
											style={{
												backgroundImage: `url(${getEmptyGemSocketIconUrl(socketColor)})`,
											}}>
											{hasChangedSocket && (
												<>
													<i className={'d-block fas fa-exclamation-circle'}></i>
												</>
											)}
										</div>,
									);
									if (hasChangedSocket && gemName)
										tippy(socketRef.value!, {
											content: (
												<>
													<strong>
														{slotName} - Socket {gemIdx + 1}
													</strong>
													<br />
													{gemName}
												</>
											),
										});
								});
							}
						}

						return <li>{itemElement}</li>;
					})}
				</ul>
			</>
		);

		trackEvent({
			action: 'settings',
			category: 'reforging',
			label: 'suggest_success',
		});
		this.reforgeDoneToast = new Toast({
			additionalClasses: ['suggest-reforges-toast'],
			variant: 'success',
			body: hasReforgeChanges ? changedReforgeMessage : <>{i18n.t('gear_tab.reforge_success.no_changes')}</>,
			autohide: !hasReforgeChanges,
			delay: 3000,
		});
	}

	onReforgeError(error: any) {
		if (isDevMode()) console.log(error);

		if (this.previousGear) this.updateGear(this.previousGear);
		trackEvent({
			action: 'settings',
			category: 'reforging',
			label: 'suggest_error',
			value: error,
		});

		new Toast({
			variant: 'error',
			body: (
				<>
					{i18n.t('sidebar.buttons.suggest_reforges.reforge_optimization_failed')}
					<p></p>
					<p>
						<b>Reason for failure:</b> <i>{error}</i>
					</p>
				</>
			),
			delay: 10000,
		});
	}

	onReforgeFinally() {
		this.progressTrackerModal.hide();

		performance.mark('reforge-optimization-end');
		const completionTimeInMs = performance.measure('reforge-optimization-measure', 'reforge-optimization-start', 'reforge-optimization-end').duration;
		if (isDevMode()) console.log('Reforge optimization took:', `${completionTimeInMs.toFixed(2)}ms`);

		trackEvent({
			action: 'settings',
			category: 'reforging',
			label: 'suggest_duration',
			value: Math.ceil(completionTimeInMs / 1000),
		});
	}

	async abortReforgeOptimization() {
		await this.sim.signalManager.abortType(RequestTypes.ReforgeOptimize);
	}

	fromProto(eventID: EventID, proto: ReforgeSettings) {
		TypedEvent.freezeAllAndDo(() => {
			this.setUseCustomEPValues(eventID, proto.useCustomEpValues);
			this.setStatCaps(eventID, Stats.fromProto(proto.statCaps));
			this.setUseSoftCapBreakpoints(eventID, proto.useSoftCapBreakpoints);
			this.setFreezeItemSlots(eventID, proto.freezeItemSlots);
			this.setFrozenItemSlots(eventID, proto.frozenItemSlots);
			this.setBreakpointLimits(eventID, Stats.fromProto(proto.breakpointLimits));
			this.setDisableUniqueGems(eventID, proto.disableUniqueGems);
			this.setMaxGemPhase(eventID, proto.maxGemPhase || Phase.Phase1);
			this.setMaxGemQuality(eventID, proto.maxGemQuality || ItemQuality.ItemQualityEpic);
		});
	}

	toProto(): ReforgeSettings {
		return ReforgeSettings.create({
			useCustomEpValues: this.useCustomEPValues,
			useSoftCapBreakpoints: this.useSoftCapBreakpoints,
			freezeItemSlots: this.freezeItemSlots,
			frozenItemSlots: [...this.frozenItemSlots],
			breakpointLimits: this.breakpointLimits.toProto(),
			statCaps: this.statCaps.toProto(),
			disableUniqueGems: this.disableUniqueGems,
			maxGemPhase: this.maxGemPhase,
			maxGemQuality: this.maxGemQuality,
		});
	}

	applyDefaults(eventID: EventID) {
		TypedEvent.freezeAllAndDo(() => {
			this.setUseCustomEPValues(eventID, false);
			this.setUseSoftCapBreakpoints(eventID, !!this.simUI.individualConfig.defaults.softCapBreakpoints?.length);
			this.setFreezeItemSlots(eventID, false);
			this.setStatCaps(eventID, this.simUI.individualConfig.defaults.statCaps || new Stats());
			this.setBreakpointLimits(eventID, this.simUI.individualConfig.defaults.breakpointLimits || new Stats());
			this.setSoftCapBreakpoints(eventID, this.simUI.individualConfig.defaults.softCapBreakpoints || []);
			this.setDisableUniqueGems(eventID, false);
			this.setMaxGemPhase(eventID, this.sim.getPhase());
			this.setMaxGemQuality(eventID, ItemQuality.ItemQualityEpic);
		});
	}
}
