import { IndividualSimUI, IndividualSimUIConfig, RaidSimPreset } from '../core/individual_sim_ui.js';
import { getSpecConfig, Player } from '../core/player.js';
import { PlayerClasses } from '../core/player_classes';
import { Spec } from '../core/proto/common.js';
import { BalanceDruidSimUI } from '../druid/balance/sim.js';
import { FeralCatDruidSimUI } from '../druid/feralcat/sim.js';
import { FeralBearDruidSimUI } from '../druid/feralbear/sim';
import { RestorationDruidSimUI } from '../druid/restoration/sim.js';
import { HunterSimUI } from '../hunter/sim.js';
import { MageSimUI } from '../mage/sim';
import { HolyPaladinSimUI } from '../paladin/holy/sim.js';
import { ProtectionPaladinSimUI } from '../paladin/protection/sim.js';
import { RetributionPaladinSimUI } from '../paladin/retribution/sim.js';
import { DisciplinePriestSimUI } from '../priest/discipline/sim';
import { HolyPriestSimUI } from '../priest/holy/sim';
import { ShadowPriestSimUI } from '../priest/shadow/sim.js';
import { RogueSimUI } from '../rogue/sim.js';
import { ElementalShamanSimUI } from '../shaman/elemental/sim.js';
import { EnhancementShamanSimUI } from '../shaman/enhancement/sim.js';
import { RestorationShamanSimUI } from '../shaman/restoration/sim.js';
import { WarlockSimUI } from '../warlock/sim.js';
import { DPSWarriorSimUI } from '../warrior/dps/sim.js';
import { ProtectionWarriorSimUI } from '../warrior/protection/sim';

export const specSimFactories: Partial<Record<Spec, (parentElem: HTMLElement, player: Player<any>) => IndividualSimUI<any>>> = {
	// Druid
	[Spec.SpecBalanceDruid]: (parentElem: HTMLElement, player: Player<any>) => new BalanceDruidSimUI(parentElem, player),
	[Spec.SpecFeralCatDruid]: (parentElem: HTMLElement, player: Player<any>) => new FeralCatDruidSimUI(parentElem, player),
	[Spec.SpecRestorationDruid]: (parentElem: HTMLElement, player: Player<any>) => new RestorationDruidSimUI(parentElem, player),
	[Spec.SpecFeralBearDruid]: (parentElem: HTMLElement, player: Player<any>) => new FeralBearDruidSimUI(parentElem, player),
	// Hunter
	[Spec.SpecHunter]: (parentElem: HTMLElement, player: Player<any>) => new HunterSimUI(parentElem, player),
	// Mage
	[Spec.SpecMage]: (parentElem: HTMLElement, player: Player<any>) => new MageSimUI(parentElem, player),
	// Paladin
	[Spec.SpecHolyPaladin]: (parentElem: HTMLElement, player: Player<any>) => new HolyPaladinSimUI(parentElem, player),
	[Spec.SpecProtectionPaladin]: (parentElem: HTMLElement, player: Player<any>) => new ProtectionPaladinSimUI(parentElem, player),
	[Spec.SpecRetributionPaladin]: (parentElem: HTMLElement, player: Player<any>) => new RetributionPaladinSimUI(parentElem, player),
	// Priest
	[Spec.SpecDisciplinePriest]: (parentElem: HTMLElement, player: Player<any>) => new DisciplinePriestSimUI(parentElem, player),
	[Spec.SpecHolyPriest]: (parentElem: HTMLElement, player: Player<any>) => new HolyPriestSimUI(parentElem, player),
	[Spec.SpecShadowPriest]: (parentElem: HTMLElement, player: Player<any>) => new ShadowPriestSimUI(parentElem, player),
	// Rogue
	[Spec.SpecRogue]: (parentElem: HTMLElement, player: Player<any>) => new RogueSimUI(parentElem, player),
	// Shaman
	[Spec.SpecElementalShaman]: (parentElem: HTMLElement, player: Player<any>) => new ElementalShamanSimUI(parentElem, player),
	[Spec.SpecEnhancementShaman]: (parentElem: HTMLElement, player: Player<any>) => new EnhancementShamanSimUI(parentElem, player),
	[Spec.SpecRestorationShaman]: (parentElem: HTMLElement, player: Player<any>) => new RestorationShamanSimUI(parentElem, player),
	// Warlock
	[Spec.SpecWarlock]: (parentElem: HTMLElement, player: Player<any>) => new WarlockSimUI(parentElem, player),
	// Warrior
	[Spec.SpecDPSWarrior]: (parentElem: HTMLElement, player: Player<any>) => new DPSWarriorSimUI(parentElem, player),
	[Spec.SpecProtectionWarrior]: (parentElem: HTMLElement, player: Player<any>) => new ProtectionWarriorSimUI(parentElem, player),
};

export const playerPresets: Array<RaidSimPreset<any>> = PlayerClasses.naturalOrder
	.map(playerClass => Object.values(playerClass.specs))
	.flat()
	.map(playerSpec => getSpecConfig(playerSpec.specID))
	.map(config => {
		const indSimUiConfig = config as IndividualSimUIConfig<any>;
		return indSimUiConfig.raidSimPresets;
	})
	.flat();

export const implementedSpecs: Array<any> = [...new Set(playerPresets.map(preset => preset.spec))];
