import clsx from 'clsx';
import tippy from 'tippy.js';
import { ref } from 'tsx-vanilla';

import i18n from '../../i18n/config.js';
import * as Mechanics from '../constants/mechanics.js';
import { IndividualSimUI } from '../individual_sim_ui';
import { Player } from '../player.js';
import { ItemSlot, PseudoStat, Race, Spec, Stat, TristateEffect, WeaponType } from '../proto/common.js';
import { ActionId } from '../proto_utils/action_id';
import { getStatName } from '../proto_utils/names.js';
import { Stats, UnitStat } from '../proto_utils/stats.js';
import { EventID, TypedEvent } from '../typed_event.js';
import { Component } from './component.js';
import { NumberPicker } from './pickers/number_picker.js';

export type StatMods = { base?: Stats; gear?: Stats; talents?: Stats; buffs?: Stats; consumes?: Stats; debuffs?: Stats; final?: Stats; stats?: Array<Stat> };
export type DisplayStat = {
	stat: UnitStat;
	notEditable?: boolean;
};

const statGroups = new Map<string, Array<DisplayStat>>([
	['Primary', [{ stat: UnitStat.fromStat(Stat.StatHealth) }, { stat: UnitStat.fromStat(Stat.StatMana) }]],
	[
		'Attributes',
		[
			{ stat: UnitStat.fromStat(Stat.StatStrength) },
			{ stat: UnitStat.fromStat(Stat.StatAgility) },
			{ stat: UnitStat.fromStat(Stat.StatStamina) },
			{ stat: UnitStat.fromStat(Stat.StatIntellect) },
			{ stat: UnitStat.fromStat(Stat.StatSpirit) },
		],
	],
	[
		'Physical',
		[
			{ stat: UnitStat.fromStat(Stat.StatAttackPower) },
			{ stat: UnitStat.fromStat(Stat.StatFeralAttackPower) },
			{ stat: UnitStat.fromStat(Stat.StatRangedAttackPower) },
			{ stat: UnitStat.fromPseudoStat(PseudoStat.PseudoStatMeleeHitPercent) },
			{ stat: UnitStat.fromPseudoStat(PseudoStat.PseudoStatMeleeCritPercent) },
			{ stat: UnitStat.fromPseudoStat(PseudoStat.PseudoStatMeleeHastePercent) },
			{ stat: UnitStat.fromPseudoStat(PseudoStat.PseudoStatRangedHitPercent) },
			{ stat: UnitStat.fromPseudoStat(PseudoStat.PseudoStatRangedCritPercent) },
			{ stat: UnitStat.fromPseudoStat(PseudoStat.PseudoStatRangedHastePercent) },
			{ stat: UnitStat.fromPseudoStat(PseudoStat.PseudoStatMeleeSpeedMultiplier) },
			{ stat: UnitStat.fromPseudoStat(PseudoStat.PseudoStatRangedSpeedMultiplier) },
			{ stat: UnitStat.fromStat(Stat.StatExpertiseRating) },
			{ stat: UnitStat.fromStat(Stat.StatArmorPenetration) },
			// { stat: UnitStat.fromPseudoStat(PseudoStat.PseudoStatBonusPhysicalDamage) },
		],
	],
	[
		'Spell',
		[
			{ stat: UnitStat.fromStat(Stat.StatSpellDamage) },
			{ stat: UnitStat.fromStat(Stat.StatArcaneDamage) },
			{ stat: UnitStat.fromStat(Stat.StatFireDamage) },
			{ stat: UnitStat.fromStat(Stat.StatFrostDamage) },
			{ stat: UnitStat.fromStat(Stat.StatHolyDamage) },
			{ stat: UnitStat.fromStat(Stat.StatNatureDamage) },
			{ stat: UnitStat.fromStat(Stat.StatShadowDamage) },
			{ stat: UnitStat.fromStat(Stat.StatHealingPower) },
			{ stat: UnitStat.fromStat(Stat.StatSpellHitRating) },
			{ stat: UnitStat.fromStat(Stat.StatSpellCritRating) },
			{ stat: UnitStat.fromPseudoStat(PseudoStat.PseudoStatCastSpeedMultiplier) },
			{ stat: UnitStat.fromStat(Stat.StatSpellPenetration) },
			{ stat: UnitStat.fromStat(Stat.StatMP5) },
		],
	],
	[
		'Defense',
		[
			{ stat: UnitStat.fromStat(Stat.StatArmor) },
			{ stat: UnitStat.fromStat(Stat.StatBonusArmor) },
			{ stat: UnitStat.fromStat(Stat.StatDefenseRating) },
			{ stat: UnitStat.fromStat(Stat.StatResilienceRating) },
			{ stat: UnitStat.fromPseudoStat(PseudoStat.PseudoStatDodgePercent) },
			{ stat: UnitStat.fromPseudoStat(PseudoStat.PseudoStatParryPercent) },
			{ stat: UnitStat.fromPseudoStat(PseudoStat.PseudoStatBlockPercent) },
			{ stat: UnitStat.fromStat(Stat.StatBlockValue) },
		],
	],
	[
		'Resistance',
		[
			{ stat: UnitStat.fromStat(Stat.StatArcaneResistance) },
			{ stat: UnitStat.fromStat(Stat.StatFireResistance) },
			{ stat: UnitStat.fromStat(Stat.StatFrostResistance) },
			{ stat: UnitStat.fromStat(Stat.StatNatureResistance) },
			{ stat: UnitStat.fromStat(Stat.StatShadowResistance) },
		],
	],
]);

export class CharacterStats extends Component {
	readonly stats: Array<UnitStat>;
	readonly valueElems: Array<HTMLTableCellElement>;
	readonly meleeCritCapValueElem: HTMLTableCellElement | undefined;
	critImmunityCapValueElem: HTMLTableCellElement | undefined;
	masteryElem: HTMLTableCellElement | undefined;
	hasRacialHitBonus = false;
	activeRacialExpertiseBonuses = [false, false];

	private readonly player: Player<any>;
	private readonly modifyDisplayStats?: (player: Player<any>) => StatMods;
	private readonly overwriteDisplayStats?: (player: Player<any>) => Required<StatMods>;

	constructor(
		parent: HTMLElement,
		simUI: IndividualSimUI<any>,
		player: CharacterStats['player'],
		statList: CharacterStats['stats'],
		modifyDisplayStats?: CharacterStats['modifyDisplayStats'],
		overwriteDisplayStats?: CharacterStats['overwriteDisplayStats'],
	) {
		super(parent, 'character-stats-root');
		this.stats = [];
		this.player = player;
		this.modifyDisplayStats = modifyDisplayStats;
		this.overwriteDisplayStats = overwriteDisplayStats;

		const label = document.createElement('label');
		label.classList.add('character-stats-label');
		label.textContent = i18n.t('sidebar.character_stats.title');
		this.rootElem.appendChild(label);

		const table = document.createElement('table');
		table.classList.add('character-stats-table');
		this.rootElem.appendChild(table);

		this.valueElems = [];
		statGroups.forEach((groupedStats, key) => {
			const filteredStats = groupedStats.filter(stat => statList.find(displayStat => displayStat.equals(stat.stat)));

			if (!filteredStats.length) return;

			const body = <tbody></tbody>;
			filteredStats.forEach(displayStat => {
				const { stat } = displayStat;
				this.stats.push(stat);

				const statName = stat.getShortName(player.getClass());
				const tableValueRef = ref<HTMLTableCellElement>();
				const row = (
					<tr className="character-stats-table-row">
						<td className="character-stats-table-label">{statName}</td>
						<td className="character-stats-table-value" ref={tableValueRef}>
							{this.bonusStatsLink(displayStat)}
						</td>
					</tr>
				);
				body.appendChild(row);

				this.valueElems.push(tableValueRef.value!);
			});

			if (key === 'Defense' && this.shouldShowCritImmunity(player)) {
				const tableValueRef = ref<HTMLTableCellElement>();
				const row = (
					<tr className="character-stats-table-row">
						<td className="character-stats-table-label">Crit Immunity</td>
						<td className="character-stats-table-value" ref={tableValueRef}></td>
					</tr>
				);

				body.appendChild(row);
				this.critImmunityCapValueElem = tableValueRef.value!;
			}
			table.appendChild(body);
		});

		if (this.shouldShowMeleeCritCap(player)) {
			const tableValueRef = ref<HTMLTableCellElement>();
			const row = (
				<tr className="character-stats-table-row">
					<td className="character-stats-table-label">Melee Crit Cap</td>
					<td className="character-stats-table-value" ref={tableValueRef}></td>
				</tr>
			);

			table.appendChild(row);
			this.meleeCritCapValueElem = tableValueRef.value!;
		}

		this.updateStats(player);
		TypedEvent.onAny([player.currentStatsEmitter, player.sim.changeEmitter, player.talentsChangeEmitter]).on(() => {
			this.updateStats(player);
		});
	}

	private updateStats(player: Player<any>) {
		const playerStats = player.getCurrentStats();
		const statMods = this.modifyDisplayStats ? this.modifyDisplayStats(this.player) : {};
		this.hasRacialHitBonus = this.player.getRace() === Race.RaceDraenei;
		this.activeRacialExpertiseBonuses = this.player.getActiveRacialExpertiseBonuses();

		const baseStats = Stats.fromProto(playerStats.baseStats);
		const gearStats = Stats.fromProto(playerStats.gearStats);
		const talentsStats = Stats.fromProto(playerStats.talentsStats);
		const buffsStats = Stats.fromProto(playerStats.buffsStats);
		const consumesStats = Stats.fromProto(playerStats.consumesStats);
		const debuffStats = CharacterStats.getDebuffStats(this.player);
		const bonusStats = player.getBonusStats();

		let finalStats = Stats.fromProto(playerStats.finalStats)
			.add(statMods.base || new Stats())
			.add(statMods.gear || new Stats())
			.add(statMods.talents || new Stats())
			.add(statMods.buffs || new Stats())
			.add(statMods.consumes || new Stats())
			.add(statMods.debuffs || new Stats())
			.add(statMods.final || new Stats())
			.add(debuffStats);

		let baseDelta = baseStats.add(statMods.base || new Stats());
		let gearDelta = gearStats
			.subtract(baseStats)
			.subtract(bonusStats)
			.add(statMods.gear || new Stats());
		let talentsDelta = talentsStats.subtract(gearStats).add(statMods.talents || new Stats());
		let buffsDelta = buffsStats.subtract(talentsStats).add(statMods.buffs || new Stats());
		let consumesDelta = consumesStats.subtract(buffsStats).add(statMods.consumes || new Stats());
		let debuffsDelta = debuffStats.add(statMods.debuffs || new Stats());

		if (this.overwriteDisplayStats) {
			const statOverwrites = this.overwriteDisplayStats(this.player);
			if (statOverwrites.stats) {
				statOverwrites.stats.forEach((stat, _) => {
					baseDelta = baseDelta.withStat(stat, statOverwrites.base.getStat(stat));
					gearDelta = gearDelta.withStat(stat, statOverwrites.gear.getStat(stat));
					talentsDelta = talentsDelta.withStat(stat, statOverwrites.talents.getStat(stat));
					buffsDelta = buffsDelta.withStat(stat, statOverwrites.buffs.getStat(stat));
					consumesDelta = consumesDelta.withStat(stat, statOverwrites.consumes.getStat(stat));
					debuffsDelta = debuffsDelta.withStat(stat, statOverwrites.debuffs.getStat(stat));
					finalStats = finalStats.withStat(stat, statOverwrites.final.getStat(stat));
				});
			}
		}

		this.stats.forEach((unitStat, idx) => {
			const bonusStatValue = unitStat.hasRootStat() ? bonusStats.getStat(unitStat.getRootStat()) : 0;
			let contextualClass: string;
			if (bonusStatValue == 0) {
				contextualClass = 'text-white';
			} else if (bonusStatValue > 0) {
				contextualClass = 'text-success';
			} else {
				contextualClass = 'text-danger';
			}

			const statLinkElemRef = ref<HTMLButtonElement>();

			const valueElem = (
				<div className="stat-value-link-container">
					<button ref={statLinkElemRef} className={clsx('stat-value-link', contextualClass)}>
						{`${this.statDisplayString(finalStats, unitStat, true)} `}
					</button>
				</div>
			);

			const statLinkElem = statLinkElemRef.value!;
			this.valueElems[idx].querySelector('.stat-value-link-container')?.remove();
			this.valueElems[idx].prepend(valueElem);

			const tooltipContent = (
				<div>
					<div className="character-stats-tooltip-row">
						<span>{i18n.t('sidebar.character_stats.tooltip.base')}</span>
						<span>{this.statDisplayString(baseDelta, unitStat, true)}</span>
					</div>
					<div className="character-stats-tooltip-row">
						<span>{i18n.t('sidebar.character_stats.tooltip.gear')}</span>
						<span>{this.statDisplayString(gearDelta, unitStat)}</span>
					</div>
					<div className="character-stats-tooltip-row">
						<span>{i18n.t('sidebar.character_stats.tooltip.talents')}</span>
						<span>{this.statDisplayString(talentsDelta, unitStat)}</span>
					</div>
					<div className="character-stats-tooltip-row">
						<span>{i18n.t('sidebar.character_stats.tooltip.buffs')}</span>
						<span>{this.statDisplayString(buffsDelta, unitStat)}</span>
					</div>
					<div className="character-stats-tooltip-row">
						<span>{i18n.t('sidebar.character_stats.tooltip.consumes')}</span>
						<span>{this.statDisplayString(consumesDelta, unitStat)}</span>
					</div>
					<div className="character-stats-tooltip-row">
						<span>{i18n.t('sidebar.character_stats.tooltip.debuffs')}</span>
						<span>{this.statDisplayString(debuffsDelta, unitStat)}</span>
					</div>
					{bonusStatValue !== 0 && (
						<div className="character-stats-tooltip-row">
							<span>{i18n.t('sidebar.character_stats.tooltip.bonus')}</span>
							<span>{this.statDisplayString(bonusStats, unitStat)}</span>
						</div>
					)}
					<div className="character-stats-tooltip-row">
						<span>{i18n.t('sidebar.character_stats.tooltip.total')}</span>
						<span>{this.statDisplayString(finalStats, unitStat, true)}</span>
					</div>
				</div>
			);

			tippy(statLinkElem, {
				content: tooltipContent,
			});
		});

		if (this.meleeCritCapValueElem) {
			const meleeCritCapInfo = player.getMeleeCritCapInfo();

			const valueElem = (
				<a href="javascript:void(0)" className="stat-value-link" attributes={{ role: 'button' }}>
					{`${this.meleeCritCapDisplayString(player, finalStats)} `}
				</a>
			);

			const capDelta = meleeCritCapInfo.playerCritCapDelta;
			if (capDelta === 0) {
				valueElem.classList.add('text-white');
			} else if (capDelta > 0) {
				valueElem.classList.add('text-danger');
			} else if (capDelta < 0) {
				valueElem.classList.add('text-success');
			}

			this.meleeCritCapValueElem.querySelector('.stat-value-link')?.remove();
			this.meleeCritCapValueElem.prepend(valueElem);

			const tooltipContent = (
				<div>
					<div className="character-stats-tooltip-row">
						<span>Glancing:</span>
						<span>{`${meleeCritCapInfo.glancing.toFixed(2)}%`}</span>
					</div>
					<div className="character-stats-tooltip-row">
						<span>Suppression:</span>
						<span>{`${meleeCritCapInfo.suppression.toFixed(2)}%`}</span>
					</div>
					<div className="character-stats-tooltip-row">
						<span>To Hit Cap:</span>
						<span>{`${meleeCritCapInfo.remainingMeleeHitCap.toFixed(2)}%`}</span>
					</div>
					<div className="character-stats-tooltip-row">
						<span>To Exp Cap:</span>
						<span>{`${meleeCritCapInfo.remainingExpertiseCap.toFixed(2)}%`}</span>
					</div>
					<div className="character-stats-tooltip-row">
						<span>Debuffs:</span>
						<span>{`${meleeCritCapInfo.debuffCrit.toFixed(2)}%`}</span>
					</div>
					{meleeCritCapInfo.specSpecificOffset != 0 && (
						<div className="character-stats-tooltip-row">
							<span>Spec Offsets:</span>
							<span>{`${meleeCritCapInfo.specSpecificOffset.toFixed(2)}%`}</span>
						</div>
					)}
					<div className="character-stats-tooltip-row">
						<span>Final Crit Cap:</span>
						<span>{`${meleeCritCapInfo.baseCritCap.toFixed(2)}%`}</span>
					</div>
					<hr />
					<div className="character-stats-tooltip-row">
						<span>Can Raise By:</span>
						<span>{`${(meleeCritCapInfo.remainingExpertiseCap + meleeCritCapInfo.remainingMeleeHitCap).toFixed(2)}%`}</span>
					</div>
				</div>
			);

			tippy(valueElem, {
				content: tooltipContent,
			});
		}

		if (this.critImmunityCapValueElem) {
			const critImmunityInfo = player.getCritImmunityInfo();

			const valueElem = (
				<a href="javascript:void(0)" className="stat-value-link" attributes={{ role: 'button' }}>
					{`${this.critImmunityCapDisplayString(player, finalStats)} `}
				</a>
			);

			const capDelta = critImmunityInfo.delta;
			if (capDelta === 0) {
				valueElem.classList.add('text-white');
			} else if (capDelta > 0) {
				valueElem.classList.add('text-danger');
			} else if (capDelta < 0) {
				valueElem.classList.add('text-success');
			}

			this.critImmunityCapValueElem.querySelector('.stat-value-link')?.remove();
			this.critImmunityCapValueElem.prepend(valueElem);

			const tooltipContent = (
				<div>
					<div className="character-stats-tooltip-row">
						<span>Defense:</span>
						<span>{`${critImmunityInfo.defense.toFixed(2)}%`}</span>
					</div>
					<div className="character-stats-tooltip-row">
						<span>Resilience:</span>
						<span>{`${critImmunityInfo.resilience.toFixed(2)}%`}</span>
					</div>
					<div className="character-stats-tooltip-row">
						<span>Total:</span>
						<span>{`${critImmunityInfo.total.toFixed(2)}%`}</span>
					</div>
				</div>
			);

			tippy(valueElem, {
				content: tooltipContent,
			});
		}
	}

	private statDisplayString(deltaStats: Stats, unitStat: UnitStat, includeBase?: boolean): string {
		const rootStat = unitStat.hasRootStat() ? unitStat.getRootStat() : null;
		let rootRatingValue = rootStat !== null ? deltaStats.getStat(rootStat) : null;
		let derivedPercentOrPointsValue = unitStat.convertDefaultUnitsToPercent(deltaStats.getUnitStat(unitStat));
		const displaySuffix = unitStat.equalsStat(Stat.StatDefenseRating) ? '' : i18n.t('sidebar.character_stats.percent_suffix');

		if (unitStat.equalsStat(Stat.StatDefenseRating) && includeBase) {
			if (rootRatingValue !== null) {
				rootRatingValue += this.player.getBaseDefense();
			}
		} else if (rootStat === Stat.StatMeleeHitRating && includeBase && this.hasRacialHitBonus) {
			// Remove the rating display and only show %
			if (rootRatingValue !== null && rootRatingValue > 0) {
				rootRatingValue -= Mechanics.PHYSICAL_HIT_RATING_PER_HIT_PERCENT;
			}
		} else if (unitStat.equalsStat(Stat.StatExpertiseRating) && includeBase) {
			const [mhWeaponExpertiseActive, ohWeaponExpertiseActive] = this.activeRacialExpertiseBonuses;

			// Remove the rating display and only show %
			if (rootRatingValue !== null && rootRatingValue > 0 && mhWeaponExpertiseActive) {
				rootRatingValue -= Mechanics.EXPERTISE_PER_QUARTER_PERCENT_REDUCTION * 4;
			}

			const matchesBothHands = mhWeaponExpertiseActive && ohWeaponExpertiseActive;
			const offHand = this.player.getEquippedItem(ItemSlot.ItemSlotOffHand);
			if (
				!matchesBothHands &&
				(mhWeaponExpertiseActive || ohWeaponExpertiseActive) &&
				offHand !== null &&
				offHand.item.weaponType !== WeaponType.WeaponTypeShield &&
				offHand.item.weaponType !== WeaponType.WeaponTypeOffHand
			) {
				const hideRootRating = rootRatingValue === null || (rootRatingValue === 0 && derivedPercentOrPointsValue !== null);
				const rootRatingString = hideRootRating ? '' : String(Math.round(rootRatingValue!));
				const mhPercentString = `${derivedPercentOrPointsValue!.toFixed(2)}` + displaySuffix;
				const ohPercentValue = derivedPercentOrPointsValue! + (ohWeaponExpertiseActive ? 1 : -1);
				const ohPercentString = `${ohPercentValue.toFixed(2)}` + displaySuffix;
				const wrappedPercentString = hideRootRating ? `${mhPercentString} / ${ohPercentString}` : ` (${mhPercentString} / ${ohPercentString})`;
				return rootRatingString + wrappedPercentString;
			}
		} else if (rootStat == Stat.StatBlockValue) {
			if (rootRatingValue !== null && rootRatingValue > 0) {
				rootRatingValue *= deltaStats.getPseudoStat(PseudoStat.PseudoStatBlockValueMultiplier) || 1;
			}
		}

		const hideRootRating = rootRatingValue === null || (rootRatingValue === 0 && derivedPercentOrPointsValue !== null);
		const rootRatingString = hideRootRating ? '' : String(Math.round(rootRatingValue!));
		const percentOrPointsString =
			derivedPercentOrPointsValue === null
				? ''
				: `${derivedPercentOrPointsValue.toFixed(unitStat.equalsStat(Stat.StatDefenseRating) ? 0 : 2)}` + displaySuffix;
		const wrappedPercentOrPointsString = hideRootRating || derivedPercentOrPointsValue === null ? percentOrPointsString : ` (${percentOrPointsString})`;
		return rootRatingString + wrappedPercentOrPointsString;
	}

	public static getDebuffStats(player: Player<any>): Stats {
		let debuffStats = new Stats();
		const debuffs = player.sim.raid.getDebuffs();
		if (debuffs.faerieFire == TristateEffect.TristateEffectImproved) {
			debuffStats = debuffStats.addPseudoStat(PseudoStat.PseudoStatMeleeHitPercent, 3);
			debuffStats = debuffStats.addPseudoStat(PseudoStat.PseudoStatRangedHitPercent, 3);
		}
		if (debuffs.improvedSealOfTheCrusader) {
			debuffStats = debuffStats.addPseudoStat(PseudoStat.PseudoStatMeleeCritPercent, 3);
			debuffStats = debuffStats.addPseudoStat(PseudoStat.PseudoStatRangedCritPercent, 3);
			debuffStats = debuffStats.addPseudoStat(PseudoStat.PseudoStatSpellCritPercent, 3);
		}
		if (debuffs.exposeWeaknessUptime && debuffs.exposeWeaknessHunterAgility) {
			debuffStats = debuffStats.addStat(Stat.StatAttackPower, debuffs.exposeWeaknessHunterAgility * 0.25);
		}

		return debuffStats;
	}

	private bonusStatsLink(displayStat: DisplayStat): HTMLElement {
		const { stat, notEditable } = displayStat;
		const rootStat = stat.getRootStat();
		const statName = getStatName(rootStat);
		const linkRef = ref<HTMLButtonElement>();
		const iconRef = ref<HTMLDivElement>();

		const link = (
			<button ref={linkRef} className={clsx('add-bonus-stats text-white ms-2', notEditable && 'd-none')} dataset={{ bsToggle: 'popover' }}>
				<i ref={iconRef} className="fas fa-plus-minus"></i>
			</button>
		);

		tippy(iconRef.value!, { content: `${i18n.t('sidebar.character_stats.bonus_prefix')} ${statName}` });
		tippy(linkRef.value!, {
			interactive: true,
			trigger: 'click',
			theme: 'bonus-stats-popover',
			placement: 'right',
			onShow: instance => {
				const picker = new NumberPicker(null, this.player, {
					id: `character-bonus-stat-${rootStat}`,
					label: `${i18n.t('sidebar.character_stats.bonus_prefix')} ${statName}`,
					extraCssClasses: ['mb-0'],
					changedEvent: (player: Player<any>) => player.bonusStatsChangeEmitter,
					getValue: (player: Player<any>) => player.getBonusStats().getStat(rootStat),
					setValue: (eventID: EventID, player: Player<any>, newValue: number) => {
						const bonusStats = player.getBonusStats().withStat(rootStat, newValue);
						player.setBonusStats(eventID, bonusStats);
						instance?.hide();
					},
				});
				instance.setContent(picker.rootElem);
			},
		});

		return link as HTMLElement;
	}

	private shouldShowMeleeCritCap(player: Player<any>): boolean {
		return player.getPlayerSpec().isMeleeDpsSpec;
	}

	private shouldShowCritImmunity(player: Player<any>): boolean {
		return player.getPlayerSpec().isTankSpec;
	}

	private critImmunityCapDisplayString(player: Player<any>, _finalStats: Stats): string {
		const critImmuneDelta = player.getCritImmunity();

		if (critImmuneDelta === 0.0) {
			return i18n.t('sidebar.character_stats.crit_cap.exact');
		}

		const prefix = critImmuneDelta > 0 ? i18n.t('sidebar.character_stats.crit_cap.under_by') : i18n.t('sidebar.character_stats.crit_cap.over_by');
		return `${prefix} ${Math.abs(critImmuneDelta).toFixed(2)}%`;
	}

	private meleeCritCapDisplayString(player: Player<any>, _finalStats: Stats): string {
		const playerCritCapDelta = player.getMeleeCritCap();

		if (playerCritCapDelta === 0.0) {
			return i18n.t('sidebar.character_stats.crit_cap.exact');
		}

		const prefix = playerCritCapDelta > 0 ? i18n.t('sidebar.character_stats.crit_cap.over_by') : i18n.t('sidebar.character_stats.crit_cap.under_by');
		return `${prefix} ${Math.abs(playerCritCapDelta).toFixed(2)}%`;
	}
}
